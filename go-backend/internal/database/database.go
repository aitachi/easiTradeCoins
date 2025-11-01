package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/easitradecoins/backend/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB    *gorm.DB
	Redis *redis.Client
)

// Config holds database configuration
type Config struct {
	MySQLDSN string
	RedisURL string
}

// InitDatabase initializes database connections
func InitDatabase(cfg *Config) error {
	var err error

	// Initialize MySQL
	DB, err = initMySQL(cfg.MySQLDSN)
	if err != nil {
		return fmt.Errorf("failed to initialize MySQL: %w", err)
	}

	// Initialize Redis (optional - for backup/cache)
	if cfg.RedisURL != "" {
		Redis, err = initRedis(cfg.RedisURL)
		if err != nil {
			log.Printf("Warning: Redis initialization failed: %v", err)
			// Don't fail if Redis is unavailable
		}
	}

	// Auto migrate tables
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// initMySQL initializes MySQL connection
func initMySQL(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// initRedis initializes Redis connection (optional)
func initRedis(redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

// autoMigrate runs database migrations
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.UserAsset{},
		&models.Order{},
		&models.Trade{},
		&models.Deposit{},
		&models.Withdrawal{},
		&models.TradingPair{},
	)
}

// Close closes all database connections
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	if Redis != nil {
		Redis.Close()
	}

	log.Println("Database connections closed")
}
