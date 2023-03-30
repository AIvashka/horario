package main

import (
	"horario/internal/config"
	"horario/internal/events"
	"horario/pkg/bot"
	"log"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to the database
	db, err := events.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Create a new bot
	b, err := bot.NewBot(cfg.BotToken, db)
	if err != nil {
		log.Fatalf("Failed to create a new bot: %v", err)
	}

	// Start the bot
	if err := b.Start(); err != nil {
		log.Fatalf("Failed to start the bot: %v", err)
	}
}
