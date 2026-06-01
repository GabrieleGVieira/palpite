package domain

import "time"

type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "PENDING"
	FriendshipStatusAccepted FriendshipStatus = "ACCEPTED"
	FriendshipStatusDeclined FriendshipStatus = "DECLINED"
	FriendshipStatusBlocked  FriendshipStatus = "BLOCKED"
)

type Friendship struct {
	ID              string
	RequesterUserID string
	AddresseeUserID string
	Status          FriendshipStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type SocialEventType string

const (
	SocialEventFriendRequestSent     SocialEventType = "FRIEND_REQUEST_SENT"
	SocialEventFriendRequestAccepted SocialEventType = "FRIEND_REQUEST_ACCEPTED"
)

type SocialEvent struct {
	ActorUserID string
	TargetID    string
	TargetType  string
	Type        SocialEventType
}
