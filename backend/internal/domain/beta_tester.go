package domain

import "time"

const (
	BetaTesterSourceLanding = "landing"

	BetaTesterStatusPending            = "pending"
	BetaTesterStatusAddedToGoogleGroup = "added_to_google_group"
	BetaTesterStatusFailed             = "failed"
)

type BetaTesterAndroid struct {
	ID           string
	Name         string
	Email        string
	Source       string
	Status       string
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
