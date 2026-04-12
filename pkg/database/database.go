package database

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect() (*gorm.DB, error) {
	dsn := buildDSN()

	logLevel := logger.Silent
	if os.Getenv("APP_ENV") == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("db.DB(): %w", err)
	}

	maxOpen, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	maxIdle, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	if maxOpen == 0 {
		maxOpen = 25
	}
	if maxIdle == 0 {
		maxIdle = 5
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return db, nil
}

func buildDSN() string {
	// Prefer DATABASE_URL (Supabase / Railway style)
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
		host, port, user, pass, name, sslmode,
	)
}
