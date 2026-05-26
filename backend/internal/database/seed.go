package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/utils"
)

const (
	apiURL = "https://api.football-data.org/v4/competitions/WC/matches?stage=GROUP_STAGE"
)

type MatchesResponse = dto.FootballDataResponse
type Match = dto.FootballDataMatch

func RunWorldCupMatchSeed() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	apiKey := os.Getenv("FOOTBALL_DATA_TOKEN")

	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	if apiKey == "" {
		log.Fatal("FOOTBALL_DATA_TOKEN not set")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	matches, err := fetchMatches(apiKey)
	if err != nil {
		log.Fatalf("failed to fetch matches: %v", err)
	}

	log.Printf("found %d matches", len(matches))

	for _, match := range matches {
		err := upsertMatch(ctx, db, match)
		if err != nil {
			log.Printf("failed to save match %d: %v", match.ID, err)
			continue
		}

		log.Printf(
			"saved: %s vs %s",
			match.HomeTeam.Name,
			match.AwayTeam.Name,
		)
	}

	log.Println("seed completed")
}

func fetchMatches(apiKey string) ([]Match, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Token", apiKey)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	var data MatchesResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Matches, nil
}

func upsertMatch(
	ctx context.Context,
	db *sql.DB,
	match Match,
) error {

	status := mapStatus(match.Status)

	query := `
		insert into world_cup_matches (
			external_id,
			home_team,
			away_team,
			stage,
			kickoff_at,
			status,
			home_score,
			away_score,
			finished_at,
			last_synced_at
		)
		values (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			now()
		)
		on conflict (external_id) where external_id is not null
		do update set
			home_team = excluded.home_team,
			away_team = excluded.away_team,
			stage = excluded.stage,
			status = excluded.status,
			home_score = excluded.home_score,
			away_score = excluded.away_score,
			finished_at = excluded.finished_at,
			last_synced_at = now()
	`

	var finishedAt *time.Time

	if status == "finished" {
		now := time.Now()
		finishedAt = &now
	}

	_, err := db.ExecContext(
		ctx,
		query,
		fmt.Sprintf("%d", match.ID),
		utils.TranslateTeam(match.HomeTeam.Name),
		utils.TranslateTeam(match.AwayTeam.Name),
		match.Stage,
		match.UTCDate,
		status,
		match.Score.FullTime.Home,
		match.Score.FullTime.Away,
		finishedAt,
	)

	return err
}

func mapStatus(apiStatus string) string {
	switch apiStatus {
	case "TIMED", "SCHEDULED":
		return "scheduled"

	case "IN_PLAY", "LIVE", "PAUSED":
		return "live"

	case "FINISHED":
		return "finished"

	case "POSTPONED":
		return "postponed"

	case "CANCELLED":
		return "cancelled"

	default:
		return "scheduled"
	}
}
