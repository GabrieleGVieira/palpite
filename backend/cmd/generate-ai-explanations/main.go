package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/ai"
	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/database"
	explanationrepo "github.com/gabrielevieira/palpitai/backend/internal/explanations/repository"
	explanationservice "github.com/gabrielevieira/palpitai/backend/internal/explanations/service"
)

func main() {
	mode := flag.String("mode", "seed", "generation mode: seed or refresh")
	fromDateArg := flag.String("from-date", "", "start date in YYYY-MM-DD")
	toDateArg := flag.String("to-date", "", "end date in YYYY-MM-DD")
	limit := flag.Int("limit", 50, "maximum matches to process")
	flag.Parse()

	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	timeoutSeconds, _ := strconv.Atoi(cfg.GeminiTimeoutSeconds)
	requestDelaySeconds, _ := strconv.Atoi(cfg.GeminiRequestDelaySeconds)
	rateLimitCooldownSeconds, _ := strconv.Atoi(cfg.GeminiRateLimitCooldownSeconds)
	rateLimitMaxWaits, _ := strconv.Atoi(cfg.GeminiRateLimitMaxWaits)
	batchSize, _ := strconv.Atoi(cfg.AIExplanationBatchSize)
	minBatchSize, _ := strconv.Atoi(cfg.AIExplanationMinBatchSize)
	maxMissingRetries, _ := strconv.Atoi(cfg.AIExplanationMaxMissingRetries)
	retryMissing, _ := strconv.ParseBool(cfg.AIExplanationRetryMissing)
	seedDays, _ := strconv.Atoi(cfg.AIExplanationSeedDays)
	refreshDays, _ := strconv.Atoi(cfg.AIExplanationRefreshDays)
	maxAgeHours, _ := strconv.Atoi(cfg.AIExplanationMaxAgeHours)

	fromDate, toDate, err := explanationWindow(*mode, *fromDateArg, *toDateArg, seedDays, refreshDays)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	aiClient, err := ai.NewGeminiClient(cfg.GeminiAPIKey, cfg.GeminiModel, time.Duration(timeoutSeconds)*time.Second)
	if err != nil {
		logger.Error("ai client setup failed", "error", err)
		os.Exit(1)
	}

	workerTimeout := time.Duration(rateLimitCooldownSeconds*rateLimitMaxWaits+requestDelaySeconds*max(*limit, 1)+600) * time.Second
	if workerTimeout < 30*time.Minute {
		workerTimeout = 30 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), workerTimeout)
	defer cancel()
	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := explanationrepo.New(db)
	service := explanationservice.NewExplanationGenerationService(repo, aiClient, cfg.GeminiModel, logger).
		WithRequestDelay(time.Duration(requestDelaySeconds)*time.Second).
		WithRateLimitCooldown(time.Duration(rateLimitCooldownSeconds)*time.Second, rateLimitMaxWaits).
		WithMissingRetry(minBatchSize, retryMissing, maxMissingRetries)
	summary, err := service.Generate(ctx, fromDate, toDate, *limit, batchSize, time.Duration(maxAgeHours)*time.Hour)
	if err != nil {
		logger.Error("ai explanation generation failed", "error", err)
		os.Exit(1)
	}

	fmt.Println("AI explanation generation finished")
	fmt.Printf("Mode: %s\n", *mode)
	fmt.Printf("From: %s\n", fromDate.Format("2006-01-02"))
	fmt.Printf("To: %s\n", toDate.Format("2006-01-02"))
	fmt.Printf("Processed: %d\n", summary.Processed)
	fmt.Printf("Generated: %d\n", summary.Generated)
	fmt.Printf("Failed: %d\n", summary.Failed)
	fmt.Printf("Rate limited: %t\n", summary.RateLimited)
	fmt.Printf("Rate limit waits: %d\n", summary.RateLimitWaits)
	fmt.Printf("Prompt version: %s\n", summary.PromptVersion)
}

func explanationWindow(mode string, fromDateArg string, toDateArg string, seedDays int, refreshDays int) (time.Time, time.Time, error) {
	if fromDateArg != "" || toDateArg != "" {
		if fromDateArg == "" || toDateArg == "" {
			return time.Time{}, time.Time{}, fmt.Errorf("--from-date and --to-date must be used together")
		}
		fromDate, err := time.Parse("2006-01-02", fromDateArg)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --from-date: %w", err)
		}
		toDate, err := time.Parse("2006-01-02", toDateArg)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --to-date: %w", err)
		}
		return fromDate, toDate, nil
	}
	if seedDays <= 0 {
		seedDays = 90
	}
	if refreshDays <= 0 {
		refreshDays = 7
	}
	now := time.Now()
	fromDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch mode {
	case "seed":
		return fromDate, fromDate.AddDate(0, 0, seedDays), nil
	case "refresh":
		return fromDate, fromDate.AddDate(0, 0, refreshDays), nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("--mode must be seed or refresh")
	}
}

func max(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
