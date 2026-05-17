package config

import "os"

type Config struct {
	DatabaseURL                 string
	Env                         string
	FootballDataAPIBaseURL      string
	FootballDataCompetitionCode string
	FootballDataSeason          string
	FootballDataToken           string
	Port                        string
	SupabaseKey                 string
	SupabaseURL                 string
}

func Load() Config {
	return Config{
		DatabaseURL:                 getEnv("DATABASE_URL", ""),
		Env:                         getEnv("APP_ENV", "development"),
		FootballDataAPIBaseURL:      getEnv("FOOTBALL_DATA_API_BASE_URL", "https://api.football-data.org/v4"),
		FootballDataCompetitionCode: getEnv("FOOTBALL_DATA_COMPETITION_CODE", "WC"),
		FootballDataSeason:          getEnv("FOOTBALL_DATA_SEASON", ""),
		FootballDataToken:           getEnv("FOOTBALL_DATA_TOKEN", ""),
		Port:                        getEnv("PORT", "3000"),
		SupabaseKey:                 getEnv("SUPABASE_KEY", ""),
		SupabaseURL:                 getEnv("SUPABASE_URL", ""),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
