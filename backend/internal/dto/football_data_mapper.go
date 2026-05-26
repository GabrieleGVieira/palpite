package dto

import (
	"strconv"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/utils"
)

func FromFootballDataMatch(match FootballDataMatch) domain.ProviderMatch {
	homeScore := match.Score.FullTime.Home
	awayScore := match.Score.FullTime.Away
	if homeScore == nil || awayScore == nil {
		homeScore = match.Score.HalfTime.Home
		awayScore = match.Score.HalfTime.Away
	}

	provider := domain.ProviderMatch{
		AwayScore:  awayScore,
		AwayTeam:   utils.TranslateTeam(footballDataTeamName(match.AwayTeam)),
		ExternalID: strconv.Itoa(match.ID),
		Goals:      make([]domain.ProviderGoal, 0, len(match.Goals)),
		HomeScore:  homeScore,
		HomeTeam:   utils.TranslateTeam(footballDataTeamName(match.HomeTeam)),
		KickoffAt:  match.UTCDate,
		Stage:      match.Stage,
		Status:     normalizeFootballDataStatus(match.Status),
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

		provider.Goals = append(provider.Goals, domain.ProviderGoal{
			AssistName: assistName,
			AwayScore:  goal.Score.Away,
			EventKey:   strings.Join(eventKeyParts, ":"),
			HomeScore:  goal.Score.Home,
			InjuryTime: goal.InjuryTime,
			Minute:     goal.Minute,
			PlayerName: goal.Scorer.Name,
			TeamName:   utils.TranslateTeam(footballDataTeamName(goal.Team)),
			Type:       goal.Type,
		})
	}

	return provider
}

func normalizeFootballDataStatus(status string) string {
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

func footballDataTeamName(team FootballDataTeam) string {
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
