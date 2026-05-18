package domain

import (
	"time"
)

type SyncSummary struct {
	ChangedMatches     int
	CreatedEvents      int
	ScoredPredictions  int
	SyncedMatches      int
	UpdatedLiveMatches int
}

type ProviderMatch struct {
	AwayScore  *int
	AwayTeam   string
	ExternalID string
	Goals      []ProviderGoal
	HomeScore  *int
	HomeTeam   string
	KickoffAt  time.Time
	Stage      string
	Status     string
}

type ProviderGoal struct {
	AssistName string
	AwayScore  *int
	EventKey   string
	HomeScore  *int
	InjuryTime *int
	Minute     int
	PlayerName string
	TeamName   string
	Type       string
}

type MatchSnapshot struct {
	AwayScore *int
	HomeScore *int
	ID        string
	Status    string
}

type AffectedGroup struct {
	ID   string
	Name string
}

type MatchDetails struct {
	AwayTeam string
	HomeTeam string
}

type GroupSummary struct {
	ID   string
	Name string
}
