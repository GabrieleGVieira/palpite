package domain

import "time"

const (
	BetaTesterSourceLanding   = "landing"
	BetaTesterPlatformAndroid = "android"

	BetaTesterStatusPending            = "pending"
	BetaTesterStatusPendingApproval    = "pending_approval"
	BetaTesterStatusAddedToGoogleGroup = "added_to_google_group"
	BetaTesterStatusApproved           = "approved"
	BetaTesterStatusRejected           = "rejected"
	BetaTesterStatusExported           = "exported"
	BetaTesterStatusFailed             = "failed"
)

type BetaTesterAndroid struct {
	ID           string
	Name         string
	Email        string
	Source       string
	Platform     string
	Status       string
	ErrorMessage string
	ApprovedAt   *time.Time
	ApprovedBy   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
