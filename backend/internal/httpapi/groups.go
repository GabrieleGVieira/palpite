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

type updateGroupRequest struct {
	Name                     string `json:"name"`
	Description              string `json:"description"`
	ParticipantLimit         *int   `json:"participant_limit"`
	HasUnlimitedParticipants bool   `json:"has_unlimited_participants"`
	IsPrivate                bool   `json:"is_private"`
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
	MemberCount          int    `json:"member_count"`
	PendingRequestsCount int    `json:"pending_requests_count"`
	Role                 string `json:"role"`
	Status               string `json:"status"`
}

type joinGroupResponse struct {
	Group            groupListItemResponse `json:"group"`
	MembershipStatus string                `json:"membership_status"`
}

type joinRequestResponse struct {
	RequestedAt time.Time `json:"requested_at"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
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
			writeError(w, http.StatusInternalServerError, "Não foi possivel listar os grupos.")
			return
		}

		writeJSON(w, http.StatusOK, map[string][]groupListItemResponse{
			"groups": groups,
		})
	}
}

func createGroupHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, displayName, err := userIDAndDisplayNameFromRequest(r, cfg)
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

		group, err := createGroup(r.Context(), db, userID, displayName, normalizedRequest)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possivel criar o grupo.")
			return
		}

		writeJSON(w, http.StatusCreated, group)
	}
}

func updateGroupHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		var request updateGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido.")
			return
		}

		normalizedRequest, err := normalizeUpdateGroupRequest(request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		group, err := updateGroup(r.Context(), db, ownerID, r.PathValue("groupID"), normalizedRequest)
		if err != nil {
			if errors.Is(err, errGroupNotFound) {
				writeError(w, http.StatusNotFound, "Grupo não encontrado.")
				return
			}

			writeError(w, http.StatusInternalServerError, "Não foi possivel atualizar o grupo.")
			return
		}

		writeJSON(w, http.StatusOK, group)
	}
}

func joinGroupHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, displayName, err := userIDAndDisplayNameFromRequest(r, cfg)
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

		response, err := joinGroup(r.Context(), db, userID, displayName, inviteCode)
		if err != nil {
			switch {
			case errors.Is(err, errGroupNotFound):
				writeError(w, http.StatusNotFound, "Grupo não encontrado.")
			case errors.Is(err, errGroupFull):
				writeError(w, http.StatusConflict, "Este grupo atingiu o limite de participantes.")
			default:
				writeError(w, http.StatusInternalServerError, "Não foi possivel entrar no grupo.")
			}
			return
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func listJoinRequestsHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		requests, err := listJoinRequests(r.Context(), db, userID, r.PathValue("groupID"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Não foi possivel listar as solicitacoes.")
			return
		}

		writeJSON(w, http.StatusOK, map[string][]joinRequestResponse{
			"requests": requests,
		})
	}
}

func approveJoinRequestHandler(cfg config.Config, db datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerID, err := userIDFromRequest(r, cfg)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		if err := approveJoinRequest(r.Context(), db, ownerID, r.PathValue("groupID"), r.PathValue("userID")); err != nil {
			switch {
			case errors.Is(err, errGroupNotFound):
				writeError(w, http.StatusNotFound, "Solicitacao não encontrada.")
			case errors.Is(err, errGroupFull):
				writeError(w, http.StatusConflict, "Este grupo atingiu o limite de participantes.")
			default:
				writeError(w, http.StatusInternalServerError, "Não foi possivel aprovar a solicitacao.")
			}
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"status": "approved",
		})
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
			gm.status,
			count(distinct all_members.user_id)::int as member_count,
			count(distinct pending_members.user_id)::int as pending_requests_count
		from groups g
		join group_members gm on gm.group_id = g.id and gm.user_id = $1 and gm.status = 'active'
		left join group_members all_members on all_members.group_id = g.id and all_members.status = 'active'
		left join group_members pending_members on pending_members.group_id = g.id and pending_members.status = 'pending'
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
			gm.role,
			gm.status
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
			&group.Status,
			&group.MemberCount,
			&group.PendingRequestsCount,
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

func joinGroup(ctx context.Context, db datastore, userID string, displayName string, inviteCode string) (joinGroupResponse, error) {
	var groupID string
	var isPrivate bool
	var participantLimit *int
	var memberCount int

	err := db.QueryRow(ctx, `
		select
			g.id,
			g.is_private,
			g.participant_limit,
			count(gm.user_id)::int as member_count
		from groups g
		left join group_members gm on gm.group_id = g.id and gm.status = 'active'
		where g.invite_code = $1
		group by g.id, g.is_private, g.participant_limit
	`, inviteCode).Scan(&groupID, &isPrivate, &participantLimit, &memberCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return joinGroupResponse{}, errGroupNotFound
	}
	if err != nil {
		return joinGroupResponse{}, err
	}

	var currentStatus string
	err = db.QueryRow(ctx, `
		select status from group_members where group_id = $1 and user_id = $2
	`, groupID, userID).Scan(&currentStatus)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return joinGroupResponse{}, err
	}
	if currentStatus == "pending" {
		group, err := groupByID(ctx, db, groupID, userID, "member", "pending")
		return joinGroupResponse{Group: group, MembershipStatus: "pending"}, err
	}
	if currentStatus == "active" {
		group, err := groupByID(ctx, db, groupID, userID, "member", "active")
		return joinGroupResponse{Group: group, MembershipStatus: "active"}, err
	}

	if participantLimit != nil && memberCount >= *participantLimit {
		return joinGroupResponse{}, errGroupFull
	}

	nextStatus := "active"
	if isPrivate {
		nextStatus = "pending"
	}

	if _, err := db.Exec(ctx, `
		insert into group_members (group_id, user_id, role, status, display_name)
		values ($1, $2, 'member', $3, $4)
		on conflict (group_id, user_id) do nothing
	`, groupID, userID, nextStatus, displayName); err != nil {
		return joinGroupResponse{}, err
	}

	group, err := groupByID(ctx, db, groupID, userID, "member", nextStatus)
	if err != nil {
		return joinGroupResponse{}, err
	}

	return joinGroupResponse{Group: group, MembershipStatus: nextStatus}, nil
}

func groupByID(ctx context.Context, db datastore, groupID string, userID string, role string, status string) (groupListItemResponse, error) {
	var group groupListItemResponse

	err := db.QueryRow(ctx, `
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
			count(distinct all_members.user_id)::int as member_count,
			count(distinct pending_members.user_id)::int as pending_requests_count
		from groups g
		left join group_members all_members on all_members.group_id = g.id and all_members.status = 'active'
		left join group_members pending_members on pending_members.group_id = g.id and pending_members.status = 'pending'
		where g.id = $1
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
			g.created_at
	`, groupID).Scan(
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
		&group.MemberCount,
		&group.PendingRequestsCount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return groupListItemResponse{}, errGroupNotFound
	}
	if err != nil {
		return groupListItemResponse{}, err
	}

	group.Role = role
	group.Status = status

	if group.OwnerID == userID {
		group.Role = "owner"
	}

	return group, nil
}

func listJoinRequests(ctx context.Context, db datastore, ownerID string, groupID string) ([]joinRequestResponse, error) {
	rows, err := db.Query(ctx, `
		select
			gm.user_id,
			gm.display_name,
			gm.joined_at
		from group_members gm
		join groups g on g.id = gm.group_id
		where gm.group_id = $1
			and g.owner_id = $2
			and gm.status = 'pending'
		order by gm.joined_at asc
	`, groupID, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []joinRequestResponse{}
	for rows.Next() {
		var request joinRequestResponse
		if err := rows.Scan(&request.UserID, &request.DisplayName, &request.RequestedAt); err != nil {
			return nil, err
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func approveJoinRequest(ctx context.Context, db datastore, ownerID string, groupID string, requesterID string) error {
	var participantLimit *int
	var memberCount int

	err := db.QueryRow(ctx, `
		select
			g.participant_limit,
			count(gm.user_id)::int as member_count
		from groups g
		left join group_members gm on gm.group_id = g.id and gm.status = 'active'
		where g.id = $1 and g.owner_id = $2
		group by g.id, g.participant_limit
	`, groupID, ownerID).Scan(&participantLimit, &memberCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return errGroupNotFound
	}
	if err != nil {
		return err
	}

	if participantLimit != nil && memberCount >= *participantLimit {
		return errGroupFull
	}

	var approvedGroupID string
	err = db.QueryRow(ctx, `
		update group_members
		set status = 'active', joined_at = now()
		where group_id = $1 and user_id = $2 and status = 'pending'
		returning group_id
	`, groupID, requesterID).Scan(&approvedGroupID)
	if errors.Is(err, pgx.ErrNoRows) {
		return errGroupNotFound
	}

	return err
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

func normalizeUpdateGroupRequest(request updateGroupRequest) (updateGroupRequest, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)

	if request.Name == "" {
		return request, errors.New("Informe o nome do grupo.")
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

func createGroup(ctx context.Context, db datastore, userID string, displayName string, request createGroupRequest) (groupResponse, error) {
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
				insert into group_members (group_id, user_id, role, display_name)
				select id, owner_id, 'owner', $2 from inserted_group
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
			displayName,
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

func updateGroup(ctx context.Context, db datastore, ownerID string, groupID string, request updateGroupRequest) (groupResponse, error) {
	var group groupResponse

	err := db.QueryRow(ctx, `
		update groups
		set
			name = $3,
			description = $4,
			participant_limit = $5,
			is_private = $6,
			updated_at = now()
		where id = $1 and owner_id = $2
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
	`, groupID, ownerID, request.Name, request.Description, request.ParticipantLimit, request.IsPrivate).Scan(
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
	if errors.Is(err, pgx.ErrNoRows) {
		return groupResponse{}, errGroupNotFound
	}
	if err != nil {
		return groupResponse{}, err
	}

	return group, nil
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
