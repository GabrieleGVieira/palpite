package main

import (
	"github.com/gabrielevieira/palpitai/backend/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	database.RunWorldCupMatchSeed()
}
