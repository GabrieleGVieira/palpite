package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/database"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		logger.Error("migration glob failed", "error", err)
		os.Exit(1)
	}
	sort.Strings(files)

	if len(files) == 0 {
		logger.Info("no migration files found")
		return
	}

	for _, file := range files {
		sql, err := os.ReadFile(file)
		if err != nil {
			logger.Error("migration read failed", "file", file, "error", err)
			os.Exit(1)
		}

		if _, err := db.Exec(ctx, string(sql)); err != nil {
			logger.Error("migration failed", "file", file, "error", err)
			os.Exit(1)
		}

		fmt.Printf("applied %s\n", file)
	}

	logger.Info("migrations finished", "count", len(files))
}
