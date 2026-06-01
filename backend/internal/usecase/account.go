package usecase

import (
	"context"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

var ErrAccountOwnsGroups = apperrors.NewConflict("account owns groups")

type AccountUsecase struct {
	db Datastore
}

func NewAccountUsecase(db Datastore) AccountUsecase {
	return AccountUsecase{db: db}
}

func (uc AccountUsecase) DeleteAccount(ctx context.Context, userID string) error {
	return DeleteAccount(ctx, uc.db, userID)
}

func (uc AccountUsecase) Profile(ctx context.Context, userID string) (dto.ProfileResponse, error) {
	return Profile(ctx, uc.db, userID)
}

func (uc AccountUsecase) UpdateProfile(ctx context.Context, userID string, request dto.UpdateProfileRequest) (dto.ProfileResponse, error) {
	return UpdateProfile(ctx, uc.db, userID, request)
}

func DeleteAccount(ctx context.Context, db Datastore, userID string) error {
	return withTx(ctx, db, func(tx repositories.Querier) error {
		ownedGroups, err := repositories.UserOwnedGroupCount(ctx, tx, userID)
		if err != nil {
			return err
		}
		if ownedGroups > 0 {
			return ErrAccountOwnsGroups
		}

		return repositories.AnonymizeAccountData(ctx, tx, userID)
	})
}

func Profile(ctx context.Context, db Datastore, userID string) (dto.ProfileResponse, error) {
	profile, err := repositories.UserProfile(ctx, db, userID)
	if err == nil {
		return profile, nil
	}
	if err == repositories.ErrNotFound {
		return dto.ProfileResponse{}, ErrGroupNotFound
	}

	return dto.ProfileResponse{}, err
}

func UpdateProfile(ctx context.Context, db Datastore, userID string, request dto.UpdateProfileRequest) (dto.ProfileResponse, error) {
	displayName := strings.TrimSpace(request.DisplayName)
	if displayName == "" {
		return dto.ProfileResponse{}, apperrors.NewValidation("Informe seu nome.")
	}

	var avatarURL *string
	if request.AvatarURL != nil {
		trimmed := strings.TrimSpace(*request.AvatarURL)
		if trimmed != "" {
			avatarURL = &trimmed
		}
	}

	isPublicProfile := true
	currentProfile, err := repositories.UserProfile(ctx, db, userID)
	if err == nil {
		isPublicProfile = currentProfile.IsPublicProfile
	} else if err != repositories.ErrNotFound {
		return dto.ProfileResponse{}, err
	}
	if request.IsPublicProfile != nil {
		isPublicProfile = *request.IsPublicProfile
	}

	profile, err := repositories.UpdateUserProfile(ctx, db, userID, displayName, avatarURL, isPublicProfile)
	if err == nil {
		return profile, nil
	}
	if err == repositories.ErrNotFound {
		return dto.ProfileResponse{}, ErrGroupNotFound
	}

	return dto.ProfileResponse{}, err
}
