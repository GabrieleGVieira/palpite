package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL                 string
	Env                         string
	FootballDataAPIBaseURL      string
	FootballDataCompetitionCode string
	FootballDataSeason          string
	FootballDataToken           string
	OpenAIAPIKey                string
	OpenAIModel                 string
	OpenAITimeoutSeconds        string
	Port                        string
	RedisURL                    string
	SupabaseKey                 string
	SupabaseURL                 string
}

func Load() Config {
	loadDotEnv()

	return Config{
		DatabaseURL:                 getEnv("DATABASE_URL", ""),
		Env:                         getEnv("APP_ENV", "development"),
		FootballDataAPIBaseURL:      getEnv("FOOTBALL_DATA_API_BASE_URL", "https://api.football-data.org/v4"),
		FootballDataCompetitionCode: getEnv("FOOTBALL_DATA_COMPETITION_CODE", "BSA"),
		FootballDataSeason:          getEnv("FOOTBALL_DATA_SEASON", ""),
		FootballDataToken:           getEnv("FOOTBALL_DATA_TOKEN", ""),
		OpenAIAPIKey:                getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:                 getEnv("OPENAI_MODEL", "gpt-4.1-mini"),
		OpenAITimeoutSeconds:        getEnv("OPENAI_TIMEOUT_SECONDS", "30"),
		Port:                        getEnv("PORT", "3000"),
		RedisURL:                    getEnv("REDIS_URL", ""),
		SupabaseKey:                 getEnv("SUPABASE_KEY", ""),
		SupabaseURL:                 getEnv("SUPABASE_URL", ""),
	}
}

func loadDotEnv() {
	_ = godotenv.Load("backend/.env")
	_ = godotenv.Load(".env")
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
