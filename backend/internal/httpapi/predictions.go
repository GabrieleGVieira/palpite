package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/jackc/pgx/v5"
)

type predictionRequest struct {
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
}

type predictionResponse struct {
	AwayScore int       `json:"away_score"`
	HomeScore int       `json:"home_score"`
	MatchID   string    `json:"match_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type matchResponse struct {
	AwayTeam     string              `json:"away_team"`
	HomeTeam     string              `json:"home_team"`
	ID           string              `json:"id"`
	KickoffAt    time.Time           `json:"kickoff_at"`
	MyPrediction *predictionResponse `json:"my_prediction"`
	Stage        string              `json:"stage"`
}

var (
	errMatchAlreadyStarted = errors.New("match already started")
	errMembershipRequired  = errors.New("active membership required")
)

func listGroupMatchesHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		matches, err := listGroupMatches(r.Context(), db, userID, r.PathValue("groupID"))
		if err != nil {
			if errors.Is(err, errMembershipRequired) {
				writeError(w, http.StatusForbidden, "Voce precisa participar deste grupo.")
				return
			}

			writeError(w, http.StatusInternalServerError, "Nao foi possivel listar os jogos.")
			return
		}

		writeJSON(w, http.StatusOK, map[string][]matchResponse{
			"matches": matches,
		})
	}
}

func savePredictionHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		var request predictionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}

		if request.HomeScore < 0 || request.HomeScore > 99 || request.AwayScore < 0 || request.AwayScore > 99 {
			writeError(w, http.StatusBadRequest, "Informe placares entre 0 e 99.")
			return
		}

		prediction, err := savePrediction(
			r.Context(),
			db,
			userID,
			r.PathValue("groupID"),
			r.PathValue("matchID"),
			request,
		)
		if err != nil {
			switch {
			case errors.Is(err, errMembershipRequired):
				writeError(w, http.StatusForbidden, "Voce precisa participar deste grupo.")
			case errors.Is(err, errMatchAlreadyStarted):
				writeError(w, http.StatusConflict, "O jogo ja comecou. Nao e mais possivel editar o palpite.")
			default:
				writeError(w, http.StatusInternalServerError, "Nao foi possivel salvar o palpite.")
			}
			return
		}

		writeJSON(w, http.StatusOK, prediction)
	}
}

func listGroupMatches(ctx context.Context, db datastore, userID string, groupID string) ([]matchResponse, error) {
	if err := ensureActiveGroupMember(ctx, db, userID, groupID); err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, `
		select
			m.id,
			m.home_team,
			m.away_team,
			m.stage,
			m.kickoff_at,
			p.home_score,
			p.away_score,
			p.updated_at
		from world_cup_matches m
		join groups g on g.id = $1
		left join predictions p on p.group_id = g.id and p.match_id = m.id and p.user_id = $2
		where g.match_scope = 'all'
			or m.home_team = any(g.selected_teams)
			or m.away_team = any(g.selected_teams)
		order by m.kickoff_at asc
	`, groupID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := []matchResponse{}
	for rows.Next() {
		var match matchResponse
		var homeScore *int
		var awayScore *int
		var updatedAt *time.Time

		if err := rows.Scan(
			&match.ID,
			&match.HomeTeam,
			&match.AwayTeam,
			&match.Stage,
			&match.KickoffAt,
			&homeScore,
			&awayScore,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		if homeScore != nil && awayScore != nil && updatedAt != nil {
			match.MyPrediction = &predictionResponse{
				AwayScore: *awayScore,
				HomeScore: *homeScore,
				MatchID:   match.ID,
				UpdatedAt: *updatedAt,
			}
		}

		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}

func savePrediction(ctx context.Context, db datastore, userID string, groupID string, matchID string, request predictionRequest) (predictionResponse, error) {
	if err := ensureActiveGroupMember(ctx, db, userID, groupID); err != nil {
		return predictionResponse{}, err
	}

	var kickoffAt time.Time
	err := db.QueryRow(ctx, `
		select m.kickoff_at
		from world_cup_matches m
		join groups g on g.id = $1
		where m.id = $2
			and (
				g.match_scope = 'all'
				or m.home_team = any(g.selected_teams)
				or m.away_team = any(g.selected_teams)
			)
	`, groupID, matchID).Scan(&kickoffAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return predictionResponse{}, errMembershipRequired
	}
	if err != nil {
		return predictionResponse{}, err
	}

	if !time.Now().UTC().Before(kickoffAt.UTC()) {
		return predictionResponse{}, errMatchAlreadyStarted
	}

	var prediction predictionResponse
	err = db.QueryRow(ctx, `
		insert into predictions (group_id, match_id, user_id, home_score, away_score)
		values ($1, $2, $3, $4, $5)
		on conflict (group_id, match_id, user_id)
		do update set
			home_score = excluded.home_score,
			away_score = excluded.away_score,
			updated_at = now()
		returning match_id, home_score, away_score, updated_at
	`, groupID, matchID, userID, request.HomeScore, request.AwayScore).Scan(
		&prediction.MatchID,
		&prediction.HomeScore,
		&prediction.AwayScore,
		&prediction.UpdatedAt,
	)
	if err != nil {
		return predictionResponse{}, err
	}

	return prediction, nil
}

func ensureActiveGroupMember(ctx context.Context, db datastore, userID string, groupID string) error {
	var exists bool
	err := db.QueryRow(ctx, `
		select exists (
			select 1
			from group_members
			where group_id = $1
				and user_id = $2
				and status = 'active'
		)
	`, groupID, userID).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return errMembershipRequired
	}

	return nil
}
