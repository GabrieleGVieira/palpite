package matchsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	defaultRequestTimeout = 10 * time.Second
	livePollInterval      = 30 * time.Second
	rateLimitGap          = 6 * time.Second
	todayPollInterval     = 5 * time.Minute
	upcomingPollInterval  = time.Hour
	upcomingWindow        = 30 * 24 * time.Hour
)

type datastore interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Publisher interface {
	Publish(ctx context.Context, event Event)
}

type Event = domain.Event

type LogPublisher struct {
	logger *slog.Logger
}

func (publisher LogPublisher) Publish(_ context.Context, event Event) {
	logger := publisher.logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("realtime event", "name", event.Name, "room", event.Room, "payload", event.Payload)
}

type Syncer struct {
	baseURL         string
	competitionCode string
	db              datastore
	httpClient      *http.Client
	inFlight        sync.Mutex
	lastRequestAt   time.Time
	logger          *slog.Logger
	publisher       Publisher
	rateMu          sync.Mutex
	season          string
	token           string
}

type Summary = domain.SyncSummary

type syncKind string

const (
	syncLive     syncKind = "live"
	syncToday    syncKind = "today"
	syncUpcoming syncKind = "upcoming"
)

type footballDataResponse = dto.FootballDataResponse
type footballDataMatch = dto.FootballDataMatch
type footballDataTeam = dto.FootballDataTeam
type providerMatch = domain.ProviderMatch
type providerGoal = domain.ProviderGoal
type matchSnapshot = domain.MatchSnapshot
type affectedGroup = domain.AffectedGroup

func New(cfg config.Config, db datastore, logger *slog.Logger) (*Syncer, bool) {
	if strings.TrimSpace(cfg.FootballDataToken) == "" {
		return nil, false
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &Syncer{
		baseURL:         strings.TrimRight(cfg.FootballDataAPIBaseURL, "/"),
		competitionCode: cfg.FootballDataCompetitionCode,
		db:              db,
		httpClient:      http.DefaultClient,
		logger:          logger,
		publisher:       LogPublisher{logger: logger},
		season:          cfg.FootballDataSeason,
		token:           cfg.FootballDataToken,
	}, true
}

func (syncer *Syncer) SetPublisher(publisher Publisher) {
	if publisher != nil {
		syncer.publisher = publisher
	}
}

func (syncer *Syncer) Run(ctx context.Context) {
	syncer.runOnce(ctx, syncUpcoming)
	syncer.runOnce(ctx, syncToday)
	syncer.runOnce(ctx, syncLive)

	liveTicker := time.NewTicker(livePollInterval)
	todayTicker := time.NewTicker(todayPollInterval)
	upcomingTicker := time.NewTicker(upcomingPollInterval)
	defer liveTicker.Stop()
	defer todayTicker.Stop()
	defer upcomingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-liveTicker.C:
			syncer.runOnce(ctx, syncLive)
		case <-todayTicker.C:
			syncer.runOnce(ctx, syncToday)
		case <-upcomingTicker.C:
			syncer.runOnce(ctx, syncUpcoming)
		}
	}
}

func (syncer *Syncer) runOnce(ctx context.Context, kind syncKind) {
	summary, err := syncer.SyncOnce(ctx, kind)
	if err != nil {
		syncer.logger.Warn("match sync failed", "kind", kind, "error", err)
		return
	}

	if summary.ChangedMatches > 0 || summary.CreatedEvents > 0 {
		syncer.logger.Info(
			"matches synced",
			"kind", kind,
			"synced_matches", summary.SyncedMatches,
			"changed_matches", summary.ChangedMatches,
			"created_events", summary.CreatedEvents,
			"updated_live_matches", summary.UpdatedLiveMatches,
			"scored_predictions", summary.ScoredPredictions,
		)
	}
}

func (syncer *Syncer) SyncOnce(ctx context.Context, kind syncKind) (Summary, error) {
	if !syncer.inFlight.TryLock() {
		return Summary{}, nil
	}
	defer syncer.inFlight.Unlock()

	shouldSync, err := syncer.shouldSync(ctx, kind)
	if err != nil {
		return Summary{}, err
	}
	if !shouldSync {
		return Summary{}, nil
	}

	matches, err := syncer.fetchMatches(ctx, kind)
	if err != nil {
		return Summary{}, err
	}

	summary := Summary{SyncedMatches: len(matches)}
	for _, match := range matches {
		match = normalizeMatch(match)
		if err := validateMatch(match); err != nil {
			syncer.logger.Warn("provider match ignored", "error", err)
			continue
		}

		snapshot, err := syncer.matchSnapshot(ctx, match)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return Summary{}, err
		}

		changedRows, matchID, err := syncer.upsertMatch(ctx, match)
		if err != nil {
			return Summary{}, err
		}

		if changedRows > 0 {
			summary.ChangedMatches += changedRows
			syncer.publishMatchChanged(ctx, snapshot, match)
		}

		createdEvents, err := syncer.syncGoals(ctx, matchID, match)
		if err != nil {
			return Summary{}, err
		}
		summary.CreatedEvents += createdEvents

		if match.HomeScore != nil && match.AwayScore != nil && (match.Status == "live" || match.Status == "finished") {
			scoredPredictions, err := syncer.scorePredictions(ctx, matchID, match)
			if err != nil {
				return Summary{}, err
			}

			if scoredPredictions > 0 && changedRows > 0 {
				if err := syncer.publishRankingChanged(ctx, matchID, match); err != nil {
					return Summary{}, err
				}
			}

			summary.ScoredPredictions += scoredPredictions
			if match.Status == "live" {
				summary.UpdatedLiveMatches++
			}
		}
	}

	return summary, nil
}

func (syncer *Syncer) fetchMatches(ctx context.Context, kind syncKind) ([]providerMatch, error) {
	if err := syncer.waitRateLimit(ctx); err != nil {
		return nil, err
	}

	endpoint, err := syncer.matchesURL(kind)
	if err != nil {
		return nil, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Auth-Token", syncer.token)

	response, err := syncer.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTooManyRequests {
		return nil, errors.New("football-data rate limit reached")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("football-data returned status %d", response.StatusCode)
	}

	var payload footballDataResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	matches := make([]providerMatch, 0, len(payload.Matches))
	for _, match := range payload.Matches {
		matches = append(matches, fromFootballData(match))
	}

	return matches, nil
}

func (syncer *Syncer) matchesURL(kind syncKind) (string, error) {
	parsedURL, err := url.Parse(syncer.baseURL + "/competitions/" + syncer.competitionCode + "/matches")
	if err != nil {
		return "", err
	}

	query := parsedURL.Query()
	if syncer.season != "" {
		query.Set("season", syncer.season)
	}

	now := time.Now().UTC()
	switch kind {
	case syncLive:
		query.Set("status", "LIVE")
	case syncToday:
		today := now.Format(time.DateOnly)
		query.Set("dateFrom", today)
		query.Set("dateTo", today)
	case syncUpcoming:
		query.Set("dateFrom", now.AddDate(0, 0, 1).Format(time.DateOnly))
		query.Set("dateTo", now.Add(upcomingWindow).Format(time.DateOnly))
	default:
		return "", fmt.Errorf("unsupported sync kind %q", kind)
	}

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

func (syncer *Syncer) waitRateLimit(ctx context.Context) error {
	syncer.rateMu.Lock()
	defer syncer.rateMu.Unlock()

	wait := rateLimitGap - time.Since(syncer.lastRequestAt)
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}

	syncer.lastRequestAt = time.Now()
	return nil
}

func (syncer *Syncer) shouldSync(ctx context.Context, kind syncKind) (bool, error) {
	if kind != syncLive {
		return true, nil
	}

	var exists bool
	err := syncer.db.QueryRow(ctx, `
		select exists (
			select 1
			from world_cup_matches
			where status = 'live'
				or (
					status not in ('finished', 'cancelled', 'postponed')
					and kickoff_at between now() - interval '3 hours' and now() + interval '30 minutes'
				)
		)
	`).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (syncer *Syncer) matchSnapshot(ctx context.Context, match providerMatch) (matchSnapshot, error) {
	var snapshot matchSnapshot
	err := syncer.db.QueryRow(ctx, `
		select id, status, home_score, away_score
		from world_cup_matches
		where home_team = $1 and away_team = $2 and kickoff_at = $3
	`, match.HomeTeam, match.AwayTeam, match.KickoffAt).Scan(
		&snapshot.ID,
		&snapshot.Status,
		&snapshot.HomeScore,
		&snapshot.AwayScore,
	)
	if err != nil {
		return matchSnapshot{}, err
	}

	return snapshot, nil
}

func (syncer *Syncer) upsertMatch(ctx context.Context, match providerMatch) (int, string, error) {
	var matchID string
	var changed bool
	err := syncer.db.QueryRow(ctx, `
		with upserted as (
			insert into world_cup_matches (
				external_id,
				home_team,
				away_team,
				stage,
				kickoff_at,
				status,
				home_score,
				away_score,
				finished_at,
				last_synced_at
			)
			values (
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				case when $6 = 'finished' then now() else null end,
				now()
			)
			on conflict (home_team, away_team, kickoff_at)
			do update set
				external_id = coalesce(excluded.external_id, world_cup_matches.external_id),
				stage = excluded.stage,
				status = excluded.status,
				home_score = excluded.home_score,
				away_score = excluded.away_score,
				finished_at = case
					when excluded.status = 'finished' then coalesce(world_cup_matches.finished_at, now())
					else null
				end,
				last_synced_at = now()
			where world_cup_matches.external_id is distinct from coalesce(excluded.external_id, world_cup_matches.external_id)
				or world_cup_matches.stage is distinct from excluded.stage
				or world_cup_matches.status is distinct from excluded.status
				or world_cup_matches.home_score is distinct from excluded.home_score
				or world_cup_matches.away_score is distinct from excluded.away_score
			returning id, true as changed
		)
		select id, changed
		from (
			select id, changed from upserted
			union all
			select id, false as changed
			from world_cup_matches
			where home_team = $2 and away_team = $3 and kickoff_at = $5
		) rows
		order by changed desc
		limit 1
	`, nullableString(match.ExternalID), match.HomeTeam, match.AwayTeam, match.Stage, match.KickoffAt, match.Status, match.HomeScore, match.AwayScore).Scan(&matchID, &changed)
	if err != nil {
		return 0, "", err
	}

	if changed {
		return 1, matchID, nil
	}

	return 0, matchID, nil
}

func (syncer *Syncer) syncGoals(ctx context.Context, matchID string, match providerMatch) (int, error) {
	created := 0
	for _, goal := range match.Goals {
		payload, err := json.Marshal(goal)
		if err != nil {
			return 0, err
		}

		commandTag, err := syncer.db.Exec(ctx, `
			insert into match_events (
				match_id,
				external_key,
				event_type,
				team_name,
				player_name,
				assist_name,
				minute,
				injury_time,
				home_score,
				away_score,
				payload
			)
			values ($1, $2, 'goal', $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
			on conflict (external_key) do nothing
		`, matchID, goal.EventKey, goal.TeamName, goal.PlayerName, goal.AssistName, goal.Minute, goal.InjuryTime, goal.HomeScore, goal.AwayScore, string(payload))
		if err != nil {
			return 0, err
		}

		if commandTag.RowsAffected() > 0 {
			created++
			syncer.publisher.Publish(ctx, Event{
				Name: "match.goal",
				Payload: map[string]any{
					"away_score":  goal.AwayScore,
					"away_team":   match.AwayTeam,
					"home_score":  goal.HomeScore,
					"home_team":   match.HomeTeam,
					"match_id":    matchID,
					"minute":      goal.Minute,
					"player_name": goal.PlayerName,
					"team_name":   goal.TeamName,
					"type":        goal.Type,
				},
				Room: "match:" + matchID,
			})
		}
	}

	return created, nil
}

func (syncer *Syncer) scorePredictions(ctx context.Context, matchID string, match providerMatch) (int, error) {
	if match.HomeScore == nil || match.AwayScore == nil {
		return 0, errors.New("cannot score predictions without match score")
	}

	commandTag, err := syncer.db.Exec(ctx, `
		update predictions p
		set
			points = case
				when p.home_score = $2 and p.away_score = $3 then 10
				when sign(p.home_score - p.away_score) = sign($2 - $3) then 5
				else 0
			end,
			scored_at = now(),
			updated_at = now()
		from world_cup_matches m
		where p.match_id = m.id
			and m.id = $1
			and p.points is distinct from case
				when p.home_score = $2 and p.away_score = $3 then 10
				when sign(p.home_score - p.away_score) = sign($2 - $3) then 5
				else 0
			end
	`, matchID, *match.HomeScore, *match.AwayScore)
	if err != nil {
		return 0, err
	}

	return int(commandTag.RowsAffected()), nil
}

func (syncer *Syncer) publishMatchChanged(ctx context.Context, previous matchSnapshot, match providerMatch) {
	payload := map[string]any{
		"away_score":      match.AwayScore,
		"away_team":       match.AwayTeam,
		"external_id":     match.ExternalID,
		"home_score":      match.HomeScore,
		"home_team":       match.HomeTeam,
		"kickoff_at":      match.KickoffAt,
		"message":         resultMessage(match.HomeTeam, match.AwayTeam, match.HomeScore, match.AwayScore),
		"previous_score":  scorePair(previous.HomeScore, previous.AwayScore),
		"previous_status": previous.Status,
		"status":          match.Status,
	}

	syncer.publisher.Publish(ctx, Event{
		Name:    "match.updated",
		Payload: payload,
		Room:    "matches",
	})

	if match.Status == "finished" {
		syncer.publisher.Publish(ctx, Event{
			Name:    "match.finished",
			Payload: payload,
			Room:    "matches",
		})
	}
}

func (syncer *Syncer) publishRankingChanged(ctx context.Context, matchID string, match providerMatch) error {
	groups, err := syncer.affectedGroups(ctx, matchID)
	if err != nil {
		return err
	}

	for _, group := range groups {
		payload := map[string]any{
			"away_score": match.AwayScore,
			"away_team":  match.AwayTeam,
			"group_id":   group.ID,
			"group_name": group.Name,
			"home_score": match.HomeScore,
			"home_team":  match.HomeTeam,
			"match_id":   matchID,
			"message":    "Ranking do grupo " + group.Name + " atualizado",
		}

		syncer.publisher.Publish(ctx, Event{
			Name:    "ranking.updated",
			Payload: payload,
			Room:    "rankings",
		})
		syncer.publisher.Publish(ctx, Event{
			Name:    "ranking.updated",
			Payload: payload,
			Room:    "group:" + group.ID,
		})
	}

	return nil
}

func (syncer *Syncer) affectedGroups(ctx context.Context, matchID string) ([]affectedGroup, error) {
	rows, err := syncer.db.Query(ctx, `
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

	groups := []affectedGroup{}
	for rows.Next() {
		var group affectedGroup
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

func fromFootballData(match footballDataMatch) providerMatch {
	homeScore := match.Score.FullTime.Home
	awayScore := match.Score.FullTime.Away
	if homeScore == nil || awayScore == nil {
		homeScore = match.Score.HalfTime.Home
		awayScore = match.Score.HalfTime.Away
	}

	provider := providerMatch{
		AwayScore:  awayScore,
		AwayTeam:   teamName(match.AwayTeam),
		ExternalID: strconv.Itoa(match.ID),
		Goals:      make([]providerGoal, 0, len(match.Goals)),
		HomeScore:  homeScore,
		HomeTeam:   teamName(match.HomeTeam),
		KickoffAt:  match.UTCDate,
		Stage:      match.Stage,
		Status:     normalizeStatus(match.Status),
	}

	for _, goal := range match.Goals {
		eventKeyParts := []string{
			provider.ExternalID,
			"goal",
			strconv.Itoa(goal.Minute),
			strconv.Itoa(goal.Team.ID),
			strconv.Itoa(goal.Scorer.ID),
			intPointerString(goal.Score.Home),
			intPointerString(goal.Score.Away),
		}

		assistName := ""
		if goal.Assist != nil {
			assistName = goal.Assist.Name
		}

		provider.Goals = append(provider.Goals, providerGoal{
			AssistName: assistName,
			AwayScore:  goal.Score.Away,
			EventKey:   strings.Join(eventKeyParts, ":"),
			HomeScore:  goal.Score.Home,
			InjuryTime: goal.InjuryTime,
			Minute:     goal.Minute,
			PlayerName: goal.Scorer.Name,
			TeamName:   teamName(goal.Team),
			Type:       goal.Type,
		})
	}

	return provider
}

func normalizeMatch(match providerMatch) providerMatch {
	match.ExternalID = strings.TrimSpace(match.ExternalID)
	match.HomeTeam = strings.TrimSpace(match.HomeTeam)
	match.AwayTeam = strings.TrimSpace(match.AwayTeam)
	match.Stage = strings.TrimSpace(match.Stage)
	match.Status = strings.ToLower(strings.TrimSpace(match.Status))

	if match.Stage == "" {
		match.Stage = "Copa do Mundo"
	}

	if match.Status == "" {
		match.Status = "scheduled"
	}

	return match
}

func normalizeStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "LIVE", "IN_PLAY", "PAUSED":
		return "live"
	case "FINISHED":
		return "finished"
	case "POSTPONED":
		return "postponed"
	case "CANCELLED", "CANCELED", "SUSPENDED":
		return "cancelled"
	default:
		return "scheduled"
	}
}

func validateMatch(match providerMatch) error {
	if match.HomeTeam == "" || match.AwayTeam == "" {
		return errors.New("home_team and away_team are required")
	}

	if match.KickoffAt.IsZero() {
		return errors.New("kickoff_at is required")
	}

	switch match.Status {
	case "scheduled", "live", "finished", "postponed", "cancelled":
	default:
		return fmt.Errorf("unsupported status %q", match.Status)
	}

	if (match.HomeScore == nil) != (match.AwayScore == nil) {
		return errors.New("home_score and away_score must be provided together")
	}

	if match.HomeScore != nil && (*match.HomeScore < 0 || *match.HomeScore > 99 || *match.AwayScore < 0 || *match.AwayScore > 99) {
		return errors.New("scores must be between 0 and 99")
	}

	return nil
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func scorePair(homeScore *int, awayScore *int) map[string]*int {
	return map[string]*int{
		"away": awayScore,
		"home": homeScore,
	}
}

func teamName(team footballDataTeam) string {
	if strings.TrimSpace(team.Name) != "" {
		return strings.TrimSpace(team.Name)
	}

	if strings.TrimSpace(team.ShortName) != "" {
		return strings.TrimSpace(team.ShortName)
	}

	return strings.TrimSpace(team.TLA)
}

func intPointerString(value *int) string {
	if value == nil {
		return "nil"
	}

	return strconv.Itoa(*value)
}

func resultMessage(homeTeam string, awayTeam string, homeScore *int, awayScore *int) string {
	if homeScore == nil || awayScore == nil {
		return homeTeam + " x " + awayTeam + " - resultado final lancado"
	}

	return fmt.Sprintf("%s %dx%d %s - resultado final lancado", homeTeam, *homeScore, *awayScore, awayTeam)
}
