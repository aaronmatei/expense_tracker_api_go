package main

import (
	"fmt"
	"log"

	"github.com/aaronmatei/expense_tracker_api_go/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("App Environment: %s\n", cfg.AppEnv)
	fmt.Printf("Port: %s\n", cfg.Port)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	//fmt.Printf("Database URL: %s\n", cfg.DatabaseURL)
	fmt.Printf("Cors Origins: %v\n", cfg.CORSOrigins)
	fmt.Printf("JWT Expiry: %v\n", cfg.JWTExpiry())
}
