package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/matchsync"
	"github.com/jackc/pgx/v5"
)

type predictionRequest struct {
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
}

type predictionResponse struct {
	AwayScore int        `json:"away_score"`
	HomeScore int        `json:"home_score"`
	MatchID   string     `json:"match_id"`
	Points    *int       `json:"points"`
	ScoredAt  *time.Time `json:"scored_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type matchResultRequest struct {
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
}

type matchResponse struct {
	AwayTeam       string              `json:"away_team"`
	FinalAwayScore *int                `json:"final_away_score"`
	FinalHomeScore *int                `json:"final_home_score"`
	FinishedAt     *time.Time          `json:"finished_at"`
	HomeTeam       string              `json:"home_team"`
	ID             string              `json:"id"`
	KickoffAt      time.Time           `json:"kickoff_at"`
	MyPrediction   *predictionResponse `json:"my_prediction"`
	Stage          string              `json:"stage"`
	Status         string              `json:"status"`
}

type rankingEntryResponse struct {
	Position    int    `json:"position"`
	TotalPoints int    `json:"total_points"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

type matchDetails struct {
	AwayTeam string
	HomeTeam string
}

type groupSummary struct {
	ID   string
	Name string
}

var (
	errMatchAlreadyStarted = errors.New("match already started")
	errMembershipRequired  = errors.New("active membership required")
)

func userScoreHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		totalScore, err := userTotalScore(r.Context(), db, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possivel carregar sua pontuacao.")
			return
		}

		writeJSON(w, http.StatusOK, map[string]int{
			"total_points": totalScore,
		})
	}
}

func groupRankingHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		ranking, err := groupRanking(r.Context(), db, userID, r.PathValue("groupID"))
		if err != nil {
			if errors.Is(err, errMembershipRequired) {
				writeError(w, http.StatusForbidden, "Você precisa participar deste grupo.")
				return
			}

			writeError(w, http.StatusInternalServerError, "Não foi possível carregar o ranking.")
			return
		}

		writeJSON(w, http.StatusOK, map[string][]rankingEntryResponse{
			"ranking": ranking,
		})
	}
}

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
				writeError(w, http.StatusForbidden, "Você precisa participar deste grupo.")
				return
			}

			writeError(w, http.StatusInternalServerError, "Não foi possível listar os jogos.")
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
				writeError(w, http.StatusForbidden, "Você precisa participar deste grupo.")
			case errors.Is(err, errMatchAlreadyStarted):
				writeError(w, http.StatusConflict, "O jogo já começou. Não é mais possível editar o palpite.")
			default:
				writeError(w, http.StatusInternalServerError, "Não foi possível salvar o palpite.")
			}
			return
		}

		writeJSON(w, http.StatusOK, prediction)
	}
}

func saveMatchResultHandler(cfg config.Config, db datastore, publisher realtimePublisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := userIDFromRequest(r, cfg); err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		var request matchResultRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}

		if request.HomeScore < 0 || request.HomeScore > 99 || request.AwayScore < 0 || request.AwayScore > 99 {
			writeError(w, http.StatusBadRequest, "Informe placares entre 0 e 99.")
			return
		}

		scoredPredictions, err := saveMatchResult(r.Context(), db, r.PathValue("matchID"), request)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possivel salvar o resultado.")
			return
		}

		if publisher != nil {
			details, _ := matchDetailsByID(r.Context(), db, r.PathValue("matchID"))
			groups, _ := groupsAffectedByMatch(r.Context(), db, r.PathValue("matchID"))
			resultMessage := formatResultMessage(details.HomeTeam, details.AwayTeam, request.HomeScore, request.AwayScore)

			publisher.Publish(r.Context(), matchsync.Event{
				Name: "match.finished",
				Payload: map[string]any{
					"away_score": request.AwayScore,
					"away_team":  details.AwayTeam,
					"home_score": request.HomeScore,
					"home_team":  details.HomeTeam,
					"match_id":   r.PathValue("matchID"),
					"message":    resultMessage,
					"status":     "finished",
				},
				Room: "matches",
			})

			if scoredPredictions > 0 {
				for _, group := range groups {
					payload := map[string]any{
						"away_score": request.AwayScore,
						"away_team":  details.AwayTeam,
						"group_id":   group.ID,
						"group_name": group.Name,
						"home_score": request.HomeScore,
						"home_team":  details.HomeTeam,
						"match_id":   r.PathValue("matchID"),
						"message":    "Ranking do grupo " + group.Name + " atualizado",
					}

					publisher.Publish(r.Context(), matchsync.Event{
						Name:    "ranking.updated",
						Payload: payload,
						Room:    "rankings",
					})
					publisher.Publish(r.Context(), matchsync.Event{
						Name:    "ranking.updated",
						Payload: payload,
						Room:    "group:" + group.ID,
					})
				}
			}
		}

		writeJSON(w, http.StatusOK, map[string]int{
			"scored_predictions": scoredPredictions,
		})
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
			m.status,
			m.kickoff_at,
			m.home_score,
			m.away_score,
			m.finished_at,
			p.home_score,
			p.away_score,
			p.points,
			p.scored_at,
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
		var finalHomeScore *int
		var finalAwayScore *int
		var finishedAt *time.Time
		var points *int
		var scoredAt *time.Time
		var updatedAt *time.Time

		if err := rows.Scan(
			&match.ID,
			&match.HomeTeam,
			&match.AwayTeam,
			&match.Stage,
			&match.Status,
			&match.KickoffAt,
			&finalHomeScore,
			&finalAwayScore,
			&finishedAt,
			&homeScore,
			&awayScore,
			&points,
			&scoredAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		match.FinalHomeScore = finalHomeScore
		match.FinalAwayScore = finalAwayScore
		match.FinishedAt = finishedAt

		if homeScore != nil && awayScore != nil && updatedAt != nil {
			match.MyPrediction = &predictionResponse{
				AwayScore: *awayScore,
				HomeScore: *homeScore,
				MatchID:   match.ID,
				Points:    points,
				ScoredAt:  scoredAt,
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

func userTotalScore(ctx context.Context, db datastore, userID string) (int, error) {
	var totalScore int
	err := db.QueryRow(ctx, `
		select coalesce(sum(p.points), 0)::int
		from predictions p
		join group_members gm on gm.group_id = p.group_id
			and gm.user_id = p.user_id
			and gm.status = 'active'
		where p.user_id = $1
	`, userID).Scan(&totalScore)
	if err != nil {
		return 0, err
	}

	return totalScore, nil
}

func groupRanking(ctx context.Context, db datastore, userID string, groupID string) ([]rankingEntryResponse, error) {
	if err := ensureActiveGroupMember(ctx, db, userID, groupID); err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, `
		with scores as (
			select
				gm.user_id,
				gm.display_name,
				coalesce(sum(p.points), 0)::int as total_points
			from group_members gm
			left join predictions p on p.group_id = gm.group_id
				and p.user_id = gm.user_id
			where gm.group_id = $1
				and gm.status = 'active'
			group by gm.user_id, gm.display_name
		)
		select
			rank() over (order by total_points desc, display_name asc)::int as position,
			user_id,
			display_name,
			total_points
		from scores
		order by position asc, display_name asc
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ranking := []rankingEntryResponse{}
	for rows.Next() {
		var entry rankingEntryResponse
		if err := rows.Scan(&entry.Position, &entry.UserID, &entry.TotalPoints); err != nil {
			return nil, err
		}

		ranking = append(ranking, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ranking, nil
}

func matchDetailsByID(ctx context.Context, db datastore, matchID string) (matchDetails, error) {
	var details matchDetails
	err := db.QueryRow(ctx, `
		select home_team, away_team
		from world_cup_matches
		where id = $1
	`, matchID).Scan(&details.HomeTeam, &details.AwayTeam)
	if err != nil {
		return matchDetails{}, err
	}

	return details, nil
}

func groupsAffectedByMatch(ctx context.Context, db datastore, matchID string) ([]groupSummary, error) {
	rows, err := db.Query(ctx, `
		select distinct g.id::text, g.name
		from groups g
		join predictions p on p.group_id = g.id
		where p.match_id = $1
		order by g.name asc
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []groupSummary{}
	for rows.Next() {
		var group groupSummary
		if err := rows.Scan(&group.ID, &group.Name); err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
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
		returning match_id, home_score, away_score, points, scored_at, updated_at
	`, groupID, matchID, userID, request.HomeScore, request.AwayScore).Scan(
		&prediction.MatchID,
		&prediction.HomeScore,
		&prediction.AwayScore,
		&prediction.Points,
		&prediction.ScoredAt,
		&prediction.UpdatedAt,
	)
	if err != nil {
		return predictionResponse{}, err
	}

	return prediction, nil
}

func saveMatchResult(ctx context.Context, db datastore, matchID string, request matchResultRequest) (int, error) {
	if _, err := db.Exec(ctx, `
		update world_cup_matches
		set
			home_score = $2,
			away_score = $3,
			status = 'finished',
			finished_at = now(),
			last_synced_at = now()
		where id = $1
	`, matchID, request.HomeScore, request.AwayScore); err != nil {
		return 0, err
	}

	commandTag, err := db.Exec(ctx, `
		update predictions
		set
			points = case
				when home_score = $2 and away_score = $3 then 10
				when sign(home_score - away_score) = sign($2 - $3) then 5
				else 0
			end,
			scored_at = now(),
			updated_at = now()
		where match_id = $1
			and points is distinct from case
				when home_score = $2 and away_score = $3 then 10
				when sign(home_score - away_score) = sign($2 - $3) then 5
				else 0
			end
	`, matchID, request.HomeScore, request.AwayScore)
	if err != nil {
		return 0, err
	}

	return int(commandTag.RowsAffected()), nil
}

func formatResultMessage(homeTeam string, awayTeam string, homeScore int, awayScore int) string {
	if homeTeam == "" || awayTeam == "" {
		return "Resultado final lancado"
	}

	return homeTeam + " " + strconv.Itoa(homeScore) + "x" + strconv.Itoa(awayScore) + " " + awayTeam + " - resultado final lancado"
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
