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
	GoogleGroupEmail               string
	GooglePrivateKey               string
	GoogleServiceAccountEmail      string
	GoogleWorkspaceDelegatedAdmin  string
	AIExplanationBatchSize         string
	AIExplanationMinBatchSize      string
	AIExplanationRetryMissing      string
	AIExplanationMaxMissingRetries string
	AIExplanationSeedDays          string
	AIExplanationRefreshDays       string
	AIExplanationMaxAgeHours       string
	Port                           string
	BetaApprovalBaseURL            string
	BetaApprovalSecret             string
	BetaAndroidPlayStoreURL        string
	PlayStoreBetaURL               string
	RedisURL                       string
	SupabaseKey                    string
	SupabaseServiceRoleKey         string
	SupabaseURL                    string
	Email                          EmailConfig
}

type EmailConfig struct {
	ResendAPIKey      string
	NotificationEmail string
	From              string
}

func Load() Config {
	loadDotEnv()

	return Config{
		DatabaseURL:                    getEnv("DATABASE_URL", ""),
		Env:                            getEnv("APP_ENV", "development"),
		FootballDataAPIBaseURL:         getEnv("FOOTBALL_DATA_API_BASE_URL", "https://api.football-data.org/v4"),
		FootballDataCompetitionCode:    getEnv("FOOTBALL_DATA_COMPETITION_CODE", "WC"),
		FootballDataSeason:             getEnv("FOOTBALL_DATA_SEASON", ""),
		FootballDataToken:              getEnv("FOOTBALL_DATA_TOKEN", ""),
		GeminiAPIKey:                   getEnv("GEMINI_API_KEY", ""),
		GeminiModel:                    getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
		GeminiRateLimitCooldownSeconds: getEnv("GEMINI_RATE_LIMIT_COOLDOWN_SECONDS", "1800"),
		GeminiRateLimitMaxWaits:        getEnv("GEMINI_RATE_LIMIT_MAX_WAITS", "1"),
		GeminiRequestDelaySeconds:      getEnv("GEMINI_REQUEST_DELAY_SECONDS", "15"),
		GeminiTimeoutSeconds:           getEnv("GEMINI_TIMEOUT_SECONDS", "30"),
		GoogleGroupEmail:               getEnv("GOOGLE_GROUP_EMAIL", ""),
		GooglePrivateKey:               getEnv("GOOGLE_PRIVATE_KEY", ""),
		GoogleServiceAccountEmail:      getEnv("GOOGLE_SERVICE_ACCOUNT_EMAIL", ""),
		GoogleWorkspaceDelegatedAdmin:  getEnv("GOOGLE_WORKSPACE_DELEGATED_ADMIN_EMAIL", ""),
		AIExplanationBatchSize:         getEnv("AI_EXPLANATION_BATCH_SIZE", "2"),
		AIExplanationMinBatchSize:      getEnv("AI_EXPLANATION_MIN_BATCH_SIZE", "1"),
		AIExplanationRetryMissing:      getEnv("AI_EXPLANATION_RETRY_MISSING", "true"),
		AIExplanationMaxMissingRetries: getEnv("AI_EXPLANATION_MAX_MISSING_RETRIES", "2"),
		AIExplanationSeedDays:          getEnv("AI_EXPLANATION_SEED_DAYS", "90"),
		AIExplanationRefreshDays:       getEnv("AI_EXPLANATION_REFRESH_DAYS", "7"),
		AIExplanationMaxAgeHours:       getEnv("AI_EXPLANATION_MAX_AGE_HOURS", "24"),
		Port:                           getEnv("PORT", "3000"),
		BetaApprovalBaseURL:            getEnv("BETA_APPROVAL_BASE_URL", ""),
		BetaApprovalSecret:             getEnv("BETA_APPROVAL_SECRET", ""),
		BetaAndroidPlayStoreURL:        getEnv("BETA_ANDROID_PLAY_STORE_URL", getEnv("PLAY_STORE_BETA_URL", "")),
		PlayStoreBetaURL:               getEnv("PLAY_STORE_BETA_URL", ""),
		RedisURL:                       getEnv("REDIS_URL", ""),
		SupabaseKey:                    getEnv("SUPABASE_KEY", ""),
		SupabaseServiceRoleKey:         getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseURL:                    getEnv("SUPABASE_URL", ""),
		Email: EmailConfig{
			ResendAPIKey:      getEnv("RESEND_API_KEY", ""),
			NotificationEmail: getEnv("BETA_SIGNUP_NOTIFICATION_EMAIL", "gabrielevieira011@gmail.com"),
			From:              getEnv("EMAIL_FROM", "Palpite! <noreply@palpite.app>"),
		},
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
