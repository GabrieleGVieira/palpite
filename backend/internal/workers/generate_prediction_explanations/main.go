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
	fromDateArg := flag.String("from-date", "", "start date in YYYY-MM-DD")
	toDateArg := flag.String("to-date", "", "end date in YYYY-MM-DD")
	limit := flag.Int("limit", 50, "maximum matches to process")
	flag.Parse()

	if *fromDateArg == "" || *toDateArg == "" {
		fmt.Fprintln(os.Stderr, "--from-date and --to-date are required")
		os.Exit(1)
	}
	fromDate, err := time.Parse("2006-01-02", *fromDateArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid --from-date:", err)
		os.Exit(1)
	}
	toDate, err := time.Parse("2006-01-02", *toDateArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid --to-date:", err)
		os.Exit(1)
	}

	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	timeoutSeconds, _ := strconv.Atoi(cfg.OpenAITimeoutSeconds)
	aiClient, err := ai.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel, time.Duration(timeoutSeconds)*time.Second)
	if err != nil {
		logger.Error("ai client setup failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := explanationrepo.New(db)
	service := explanationservice.NewExplanationGenerationService(repo, aiClient, cfg.OpenAIModel, logger)
	summary, err := service.Generate(ctx, fromDate, toDate, *limit)
	if err != nil {
		logger.Error("ai explanation generation failed", "error", err)
		os.Exit(1)
	}

	fmt.Println("AI explanation generation finished")
	fmt.Printf("Processed: %d\n", summary.Processed)
	fmt.Printf("Generated: %d\n", summary.Generated)
	fmt.Printf("Skipped: %d\n", summary.Skipped)
	fmt.Printf("Failed: %d\n", summary.Failed)
	fmt.Printf("Prompt version: %s\n", summary.PromptVersion)
}
