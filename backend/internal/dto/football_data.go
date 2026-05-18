package dto

import (
	"encoding/json"
	"time"
)

type FootballDataResponse struct {
	Matches []FootballDataMatch `json:"matches"`
}

type FootballDataMatch struct {
	AwayTeam      FootballDataTeam   `json:"awayTeam"`
	Bookings      []json.RawMessage  `json:"bookings"`
	Goals         []FootballDataGoal `json:"goals"`
	Group         *string            `json:"group"`
	HomeTeam      FootballDataTeam   `json:"homeTeam"`
	ID            int                `json:"id"`
	LastUpdated   string             `json:"lastUpdated"`
	Matchday      *int               `json:"matchday"`
	Minute        *string            `json:"minute"`
	Penalties     []json.RawMessage  `json:"penalties"`
	Score         FootballDataScore  `json:"score"`
	Stage         string             `json:"stage"`
	Status        string             `json:"status"`
	Substitutions []json.RawMessage  `json:"substitutions"`
	UTCDate       time.Time          `json:"utcDate"`
	Venue         *string            `json:"venue"`
}

type FootballDataTeam struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	TLA       string `json:"tla"`
}

type FootballDataScore struct {
	Duration string             `json:"duration"`
	FullTime FootballDataResult `json:"fullTime"`
	HalfTime FootballDataResult `json:"halfTime"`
	Winner   *string            `json:"winner"`
}

type FootballDataResult struct {
	Away *int `json:"away"`
	Home *int `json:"home"`
}

type FootballDataGoal struct {
	Assist     *FootballDataPerson `json:"assist"`
	InjuryTime *int                `json:"injuryTime"`
	Minute     int                 `json:"minute"`
	Score      FootballDataResult  `json:"score"`
	Scorer     FootballDataPerson  `json:"scorer"`
	Team       FootballDataTeam    `json:"team"`
	Type       string              `json:"type"`
}

type FootballDataPerson struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
