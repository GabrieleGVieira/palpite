package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

const (
	FeedEventMemberJoined  = "member_joined"
	FeedEventLeaderChanged = "leader_changed"
	FeedEventExactScore    = "exact_score"
	FeedEventMatchFinished = "match_finished"
	FeedEventTop3Reached   = "top3_reached"
)

var (
	ErrInvalidReactionType = apperrors.NewValidation("invalid reaction type")
)

type FeedUsecase struct {
	db Datastore
}

func NewFeedUsecase(db Datastore) FeedUsecase {
	return FeedUsecase{db: db}
}

func (uc FeedUsecase) List(ctx context.Context, userID string, groupID string, page int, pageSize int) (dto.GroupFeedResponse, error) {
	return ListFeed(ctx, uc.db, userID, groupID, page, pageSize)
}

func (uc FeedUsecase) React(ctx context.Context, userID string, groupID string, eventID string, reactionType string) error {
	return ReactToFeedEvent(ctx, uc.db, userID, groupID, eventID, reactionType)
}

func (uc FeedUsecase) DeleteReaction(ctx context.Context, userID string, groupID string, eventID string, reactionType string) error {
	return DeleteFeedReaction(ctx, uc.db, userID, groupID, eventID, reactionType)
}

type FeedEventService struct {
	db     Datastore
	logger *slog.Logger
}

func NewFeedEventService(db Datastore, logger *slog.Logger) FeedEventService {
	if logger == nil {
		logger = slog.Default()
	}

	return FeedEventService{db: db, logger: logger}
}

func (service FeedEventService) Publish(ctx context.Context, input repositories.FeedEventInput) error {
	id, created, err := repositories.InsertFeedEvent(ctx, service.db, input)
	if err != nil {
		return err
	}
	if created {
		service.logger.Info("feed event created", "id", id, "group_id", input.GroupID, "event_type", input.EventType)
	}

	return nil
}

func (service FeedEventService) PublishMany(ctx context.Context, events []repositories.FeedEventInput) error {
	for _, event := range events {
		if err := service.Publish(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func ListFeed(ctx context.Context, db Datastore, userID string, groupID string, page int, pageSize int) (dto.GroupFeedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	if err := EnsureActiveGroupMember(ctx, db, userID, groupID); err != nil {
		return dto.GroupFeedResponse{}, err
	}

	slog.Info("feed fetch requested", "group_id", groupID, "user_id", userID, "page", page, "page_size", pageSize)
	events, hasMore, err := repositories.ListFeedEvents(ctx, db, groupID, userID, page, pageSize)
	if err != nil {
		return dto.GroupFeedResponse{}, err
	}
	slog.Info("feed fetch completed", "group_id", groupID, "user_id", userID, "page", page, "count", len(events), "has_more", hasMore)

	return dto.GroupFeedResponse{
		Events:  events,
		HasMore: hasMore,
		Page:    page,
	}, nil
}

func ReactToFeedEvent(ctx context.Context, db Datastore, userID string, groupID string, eventID string, reactionType string) error {
	if !isValidFeedReaction(reactionType) {
		return ErrInvalidReactionType
	}
	if err := ensureFeedReactionAllowed(ctx, db, userID, groupID, eventID); err != nil {
		return err
	}

	return repositories.UpsertFeedReaction(ctx, db, groupID, eventID, userID, reactionType)
}

func DeleteFeedReaction(ctx context.Context, db Datastore, userID string, groupID string, eventID string, reactionType string) error {
	if reactionType != "" && !isValidFeedReaction(reactionType) {
		return ErrInvalidReactionType
	}
	if err := ensureFeedReactionAllowed(ctx, db, userID, groupID, eventID); err != nil {
		return err
	}

	return repositories.DeleteFeedReaction(ctx, db, groupID, eventID, userID, reactionType)
}

func ensureFeedReactionAllowed(ctx context.Context, db Datastore, userID string, groupID string, eventID string) error {
	if err := EnsureActiveGroupMember(ctx, db, userID, groupID); err != nil {
		return err
	}

	exists, err := repositories.FeedEventExists(ctx, db, groupID, eventID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrGroupNotFound
	}

	return nil
}

func isValidFeedReaction(reactionType string) bool {
	switch reactionType {
	case "clap", "fire", "laugh", "surprised", "target":
		return true
	default:
		return false
	}
}

func PublishRankingFeedEvents(ctx context.Context, db repositories.Querier, groupID string, matchID string, before []repositories.RankingSnapshot, after []repositories.RankingSnapshot) error {
	beforeByUser := map[string]repositories.RankingSnapshot{}
	for _, entry := range before {
		beforeByUser[entry.UserID] = entry
	}

	var previousLeaderID string
	if len(before) > 0 {
		previousLeaderID = before[0].UserID
	}
	for _, entry := range after {
		if entry.Position == 1 && entry.UserID != previousLeaderID {
			actorID := entry.UserID
			if _, _, err := repositories.InsertFeedEvent(ctx, db, repositories.FeedEventInput{
				ActorUserID: &actorID,
				EventType:   FeedEventLeaderChanged,
				GroupID:     groupID,
				MatchID:     &matchID,
				Metadata: map[string]any{
					"position": entry.Position,
				},
			}); err != nil {
				return err
			}
		}

		beforeEntry, hadBefore := beforeByUser[entry.UserID]
		if entry.Position <= 3 && (!hadBefore || beforeEntry.Position > 3) {
			actorID := entry.UserID
			if _, _, err := repositories.InsertFeedEvent(ctx, db, repositories.FeedEventInput{
				ActorUserID: &actorID,
				EventType:   FeedEventTop3Reached,
				GroupID:     groupID,
				MatchID:     &matchID,
				Metadata: map[string]any{
					"position": entry.Position,
				},
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func PublishMatchScoringFeedEvents(ctx context.Context, db repositories.Querier, matchID string, homeScore int, awayScore int, beforeByGroup map[string][]repositories.RankingSnapshot) error {
	finishedEvents, err := repositories.GroupsForFinishedMatch(ctx, db, matchID)
	if err != nil {
		return err
	}
	for _, event := range finishedEvents {
		if _, _, err := repositories.InsertFeedEvent(ctx, db, event); err != nil {
			return err
		}
	}

	exactScoreEvents, err := repositories.ExactScorePredictionsForMatch(ctx, db, matchID, homeScore, awayScore)
	if err != nil {
		return err
	}
	for _, event := range exactScoreEvents {
		if _, _, err := repositories.InsertFeedEvent(ctx, db, event); err != nil {
			return err
		}
	}

	for groupID, before := range beforeByGroup {
		after, err := repositories.GroupRankingSnapshot(ctx, db, groupID)
		if err != nil {
			return err
		}
		if err := PublishRankingFeedEvents(ctx, db, groupID, matchID, before, after); err != nil {
			return err
		}
	}

	return nil
}

func RankingSnapshotsByAffectedGroup(ctx context.Context, db repositories.Querier, matchID string) (map[string][]repositories.RankingSnapshot, error) {
	groups, err := repositories.AffectedGroupsByMatch(ctx, db, matchID)
	if err != nil {
		return nil, err
	}

	snapshots := map[string][]repositories.RankingSnapshot{}
	for _, group := range groups {
		ranking, err := repositories.GroupRankingSnapshot(ctx, db, group.ID)
		if err != nil {
			return nil, err
		}
		snapshots[group.ID] = ranking
	}

	return snapshots, nil
}

func mapRepositoryNotFound(err error) error {
	if errors.Is(err, repositories.ErrNotFound) {
		return ErrGroupNotFound
	}

	return err
}
