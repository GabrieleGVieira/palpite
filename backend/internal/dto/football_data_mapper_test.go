package dto

import (
	"testing"
	"time"
)

func TestFromFootballDataMatchTranslatesTeams(t *testing.T) {
	match := FromFootballDataMatch(FootballDataMatch{
		AwayTeam: FootballDataTeam{Name: "Morocco"},
		Goals: []FootballDataGoal{
			{
				Minute: 10,
				Scorer: FootballDataPerson{
					ID:   1,
					Name: "Player",
				},
				Score: FootballDataResult{Home: intPointer(1), Away: intPointer(0)},
				Team:  FootballDataTeam{Name: "Brazil"},
			},
		},
		HomeTeam: FootballDataTeam{Name: "Brazil"},
		ID:       123,
		Status:   "TIMED",
		UTCDate:  time.Date(2026, 6, 1, 18, 0, 0, 0, time.UTC),
	})

	if match.HomeTeam != "Brasil" {
		t.Fatalf("expected translated home team, got %q", match.HomeTeam)
	}
	if match.AwayTeam != "Marrocos" {
		t.Fatalf("expected translated away team, got %q", match.AwayTeam)
	}
	if match.Goals[0].TeamName != "Brasil" {
		t.Fatalf("expected translated goal team, got %q", match.Goals[0].TeamName)
	}
}

func intPointer(value int) *int {
	return &value
}
