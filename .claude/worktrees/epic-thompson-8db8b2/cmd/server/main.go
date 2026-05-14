package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aaronmatei/expense_tracker_api_go/internal/config"
	"github.com/aaronmatei/expense_tracker_api_go/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	//Database connection
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("Hello, expense tracker!")
	fmt.Printf("Environment: %s\n", cfg.AppEnv)
	fmt.Printf("Connected to Postgres pool (max %d conns)\n", pool.Config().MaxConns)
}
