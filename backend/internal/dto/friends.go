package dto

import "time"

type FriendRequest struct {
	UserID string `json:"userId"`
}

type FriendResponse struct {
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UserID    string    `json:"userId"`
}

type PendingFriendRequestResponse struct {
	AvatarURL    *string   `json:"avatarUrl,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	FriendshipID string    `json:"id"`
	Name         string    `json:"name"`
	RequesterID  string    `json:"userId"`
}

type UserSearchResponse struct {
	AvatarURL        *string `json:"avatarUrl,omitempty"`
	FriendshipStatus *string `json:"friendshipStatus"`
	ID               string  `json:"id"`
	Name             string  `json:"name"`
}

type PublicProfileResponse struct {
	AvatarURL        *string    `json:"avatarUrl,omitempty"`
	GlobalRanking    *int       `json:"globalRanking"`
	GroupsCount      int        `json:"groupsCount"`
	JoinedAt         *time.Time `json:"joinedAt"`
	Name             string     `json:"name"`
	PredictionsCount int        `json:"predictionsCount"`
	TotalPoints      int        `json:"totalPoints"`
	UserID           string     `json:"userId"`
}
