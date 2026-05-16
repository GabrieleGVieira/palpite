package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const inviteCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

type createGroupRequest struct {
	Name                     string   `json:"name"`
	Description              string   `json:"description"`
	MatchScope               string   `json:"match_scope"`
	SelectedTeams            []string `json:"selected_teams"`
	ParticipantLimit         *int     `json:"participant_limit"`
	HasUnlimitedParticipants bool     `json:"has_unlimited_participants"`
	IsPrivate                bool     `json:"is_private"`
}

type joinGroupRequest struct {
	InviteCode string `json:"invite_code"`
}

type groupResponse struct {
	ID               string    `json:"id"`
	OwnerID          string    `json:"owner_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	MatchScope       string    `json:"match_scope"`
	SelectedTeams    []string  `json:"selected_teams"`
	ParticipantLimit *int      `json:"participant_limit"`
	IsPrivate        bool      `json:"is_private"`
	InviteCode       string    `json:"invite_code"`
	CreatedAt        time.Time `json:"created_at"`
}

type groupListItemResponse struct {
	groupResponse
	MemberCount int    `json:"member_count"`
	Role        string `json:"role"`
}

func listGroupsHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		groups, err := listGroups(r.Context(), db, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Nao foi possivel listar os grupos.")
			return
		}

		writeJSON(w, http.StatusOK, map[string][]groupListItemResponse{
			"groups": groups,
		})
	}
}

func createGroupHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		var request createGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}

		normalizedRequest, err := normalizeCreateGroupRequest(request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		group, err := createGroup(r.Context(), db, userID, normalizedRequest)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Nao foi possivel criar o grupo.")
			return
		}

		writeJSON(w, http.StatusCreated, group)
	}
}

func joinGroupHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		var request joinGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}

		inviteCode := normalizeInviteCode(request.InviteCode)
		if inviteCode == "" {
			writeError(w, http.StatusBadRequest, "Informe o codigo do grupo.")
			return
		}

		group, err := joinGroup(r.Context(), db, userID, inviteCode)
		if err != nil {
			switch {
			case errors.Is(err, errGroupNotFound):
				writeError(w, http.StatusNotFound, "Grupo nao encontrado.")
			case errors.Is(err, errGroupFull):
				writeError(w, http.StatusConflict, "Este grupo atingiu o limite de participantes.")
			default:
				writeError(w, http.StatusInternalServerError, "Nao foi possivel entrar no grupo.")
			}
			return
		}

		writeJSON(w, http.StatusOK, group)
	}
}

var (
	errGroupFull     = errors.New("group is full")
	errGroupNotFound = errors.New("group not found")
)

func listGroups(ctx context.Context, db datastore, userID string) ([]groupListItemResponse, error) {
	rows, err := db.Query(ctx, `
		select
			g.id,
			g.owner_id,
			g.name,
			g.description,
			g.match_scope,
			g.selected_teams,
			g.participant_limit,
			g.is_private,
			g.invite_code,
			g.created_at,
			gm.role,
			count(all_members.user_id)::int as member_count
		from groups g
		join group_members gm on gm.group_id = g.id and gm.user_id = $1
		left join group_members all_members on all_members.group_id = g.id
		group by
			g.id,
			g.owner_id,
			g.name,
			g.description,
			g.match_scope,
			g.selected_teams,
			g.participant_limit,
			g.is_private,
			g.invite_code,
			g.created_at,
			gm.role
		order by g.created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []groupListItemResponse{}
	for rows.Next() {
		var group groupListItemResponse
		if err := rows.Scan(
			&group.ID,
			&group.OwnerID,
			&group.Name,
			&group.Description,
			&group.MatchScope,
			&group.SelectedTeams,
			&group.ParticipantLimit,
			&group.IsPrivate,
			&group.InviteCode,
			&group.CreatedAt,
			&group.Role,
			&group.MemberCount,
		); err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func joinGroup(ctx context.Context, db datastore, userID string, inviteCode string) (groupListItemResponse, error) {
	var groupID string
	var participantLimit *int
	var memberCount int

	err := db.QueryRow(ctx, `
		select
			g.id,
			g.participant_limit,
			count(gm.user_id)::int as member_count
		from groups g
		left join group_members gm on gm.group_id = g.id
		where g.invite_code = $1
		group by g.id, g.participant_limit
	`, inviteCode).Scan(&groupID, &participantLimit, &memberCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return groupListItemResponse{}, errGroupNotFound
	}
	if err != nil {
		return groupListItemResponse{}, err
	}

	if participantLimit != nil && memberCount >= *participantLimit {
		var alreadyMember bool
		if err := db.QueryRow(ctx, `
			select exists (
				select 1 from group_members where group_id = $1 and user_id = $2
			)
		`, groupID, userID).Scan(&alreadyMember); err != nil {
			return groupListItemResponse{}, err
		}

		if !alreadyMember {
			return groupListItemResponse{}, errGroupFull
		}
	}

	if _, err := db.Exec(ctx, `
		insert into group_members (group_id, user_id, role)
		values ($1, $2, 'member')
		on conflict (group_id, user_id) do nothing
	`, groupID, userID); err != nil {
		return groupListItemResponse{}, err
	}

	groups, err := listGroups(ctx, db, userID)
	if err != nil {
		return groupListItemResponse{}, err
	}

	for _, group := range groups {
		if group.ID == groupID {
			return group, nil
		}
	}

	return groupListItemResponse{}, errGroupNotFound
}

func normalizeCreateGroupRequest(request createGroupRequest) (createGroupRequest, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)
	request.MatchScope = strings.TrimSpace(request.MatchScope)

	if request.Name == "" {
		return request, errors.New("Informe o nome do grupo.")
	}

	if request.MatchScope != "all" && request.MatchScope != "selected" {
		return request, errors.New("Informe uma abrangencia de jogos valida.")
	}

	if request.MatchScope == "all" {
		request.SelectedTeams = []string{}
	}

	if request.MatchScope == "selected" {
		request.SelectedTeams = normalizeTeams(request.SelectedTeams)
		if len(request.SelectedTeams) == 0 {
			return request, errors.New("Selecione pelo menos uma selecao.")
		}
	}

	if request.HasUnlimitedParticipants {
		request.ParticipantLimit = nil
	} else if request.ParticipantLimit == nil || *request.ParticipantLimit < 2 {
		return request, errors.New("O limite precisa ser maior que 1.")
	}

	return request, nil
}

func normalizeInviteCode(inviteCode string) string {
	inviteCode = strings.TrimSpace(inviteCode)
	inviteCode = strings.ToUpper(inviteCode)
	inviteCode = strings.ReplaceAll(inviteCode, " ", "")
	inviteCode = strings.ReplaceAll(inviteCode, "-", "")

	return inviteCode
}

func normalizeTeams(teams []string) []string {
	seen := map[string]bool{}
	normalizedTeams := make([]string, 0, len(teams))

	for _, team := range teams {
		team = strings.TrimSpace(team)
		if team == "" || seen[team] {
			continue
		}

		seen[team] = true
		normalizedTeams = append(normalizedTeams, team)
	}

	return normalizedTeams
}

func createGroup(ctx context.Context, db datastore, userID string, request createGroupRequest) (groupResponse, error) {
	var group groupResponse

	for range 5 {
		inviteCode, err := generateInviteCode()
		if err != nil {
			return group, err
		}

		err = db.QueryRow(ctx, `
			with inserted_group as (
				insert into groups (
					owner_id,
					name,
					description,
					match_scope,
					selected_teams,
					participant_limit,
					is_private,
					invite_code
				)
				values ($1, $2, $3, $4, $5, $6, $7, $8)
				returning
					id,
					owner_id,
					name,
					description,
					match_scope,
					selected_teams,
					participant_limit,
					is_private,
					invite_code,
					created_at
			),
			inserted_member as (
				insert into group_members (group_id, user_id, role)
				select id, owner_id, 'owner' from inserted_group
			)
			select
				id,
				owner_id,
				name,
				description,
				match_scope,
				selected_teams,
				participant_limit,
				is_private,
				invite_code,
				created_at
			from inserted_group
		`,
			userID,
			request.Name,
			request.Description,
			request.MatchScope,
			request.SelectedTeams,
			request.ParticipantLimit,
			request.IsPrivate,
			inviteCode,
		).Scan(
			&group.ID,
			&group.OwnerID,
			&group.Name,
			&group.Description,
			&group.MatchScope,
			&group.SelectedTeams,
			&group.ParticipantLimit,
			&group.IsPrivate,
			&group.InviteCode,
			&group.CreatedAt,
		)
		if err == nil {
			return group, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return group, err
		}

		return group, err
	}

	return group, errors.New("failed to generate unique invite code")
}

func generateInviteCode() (string, error) {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	for index, value := range buffer {
		buffer[index] = inviteCodeAlphabet[int(value)%len(inviteCodeAlphabet)]
	}

	return string(buffer), nil
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
