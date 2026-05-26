package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL                    string
	Env                            string
	FootballDataAPIBaseURL         string
	FootballDataCompetitionCode    string
	FootballDataSeason             string
	FootballDataToken              string
	GeminiAPIKey                   string
	GeminiModel                    string
	GeminiRateLimitCooldownSeconds string
	GeminiRateLimitMaxWaits        string
	GeminiRequestDelaySeconds      string
	GeminiTimeoutSeconds           string
	Port                           string
	RedisURL                       string
	SupabaseKey                    string
	SupabaseURL                    string
}

func Load() Config {
	loadDotEnv()

	return Config{
		DatabaseURL:                    getEnv("DATABASE_URL", ""),
		Env:                            getEnv("APP_ENV", "development"),
		FootballDataAPIBaseURL:         getEnv("FOOTBALL_DATA_API_BASE_URL", "https://api.football-data.org/v4"),
		FootballDataCompetitionCode:    getEnv("FOOTBALL_DATA_COMPETITION_CODE", "BSA"),
		FootballDataSeason:             getEnv("FOOTBALL_DATA_SEASON", ""),
		FootballDataToken:              getEnv("FOOTBALL_DATA_TOKEN", ""),
		GeminiAPIKey:                   getEnv("GEMINI_API_KEY", ""),
		GeminiModel:                    getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
		GeminiRateLimitCooldownSeconds: getEnv("GEMINI_RATE_LIMIT_COOLDOWN_SECONDS", "1800"),
		GeminiRateLimitMaxWaits:        getEnv("GEMINI_RATE_LIMIT_MAX_WAITS", "1"),
		GeminiRequestDelaySeconds:      getEnv("GEMINI_REQUEST_DELAY_SECONDS", "15"),
		GeminiTimeoutSeconds:           getEnv("GEMINI_TIMEOUT_SECONDS", "30"),
		Port:                           getEnv("PORT", "3000"),
		RedisURL:                       getEnv("REDIS_URL", ""),
		SupabaseKey:                    getEnv("SUPABASE_KEY", ""),
		SupabaseURL:                    getEnv("SUPABASE_URL", ""),
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
