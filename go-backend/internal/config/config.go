//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	MySQL       MySQLConfig
	Redis       RedisConfig
	Kafka       KafkaConfig
	Elasticsearch ElasticsearchConfig
	Security    SecurityConfig
	Blockchain  BlockchainConfig
	Risk        RiskConfig
	Features    FeatureFlags
	Monitoring  MonitoringConfig
}

// AppConfig contains application-level configuration
type AppConfig struct {
	Env         string
	Port        string
	MetricsPort string
	Version     string
}

// DatabaseConfig for PostgreSQL
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConnections  int
	MaxIdleConns    int
}

// MySQLConfig for MySQL database
type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// RedisConfig for Redis cache
type RedisConfig struct {
	Host       string
	Port       string
	Password   string
	DB         int
	MaxRetries int
	PoolSize   int
}

// KafkaConfig for Kafka message queue
type KafkaConfig struct {
	Brokers             []string
	TopicOrders         string
	TopicTrades         string
	TopicNotifications  string
	TopicRiskEvents     string
	TopicMarketData     string
	GroupID             string
}

// ElasticsearchConfig for Elasticsearch
type ElasticsearchConfig struct {
	Host   string
	Port   string
	Scheme string
}

// SecurityConfig for authentication and encryption
type SecurityConfig struct {
	JWTSecret      string
	JWTExpiration  time.Duration
	BCryptCost     int
}

// BlockchainConfig for Ethereum/Sepolia interaction
type BlockchainConfig struct {
	RPCURL                      string
	ChainID                     int64
	PrivateKey                  string
	ContractAddressStaking      string
	ContractAddressAirdrop      string
	ContractAddressTokenFactory string
	ContractAddressDEXAggregator string
	ContractAddressLiquidityMining string
}

// RiskConfig for risk management parameters
type RiskConfig struct {
	OrderRateLimit            int
	OrderRateWindow           time.Duration
	MaxPriceDeviation         float64
	WithdrawalDailyLimit      float64
	LargeWithdrawalThreshold  float64
}

// FeatureFlags for enabling/disabling features
type FeatureFlags struct {
	EnableWebSocket        bool
	EnableStopOrderMonitor bool
	EnableRiskManager      bool
	EnableSwagger          bool
}

// MonitoringConfig for observability
type MonitoringConfig struct {
	PrometheusEnabled bool
	LogLevel          string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Env:         getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8080"),
			MetricsPort: getEnv("APP_METRICS_PORT", "8081"),
			Version:     getEnv("BUILD_VERSION", "latest"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "socialfi"),
			Password:        getEnv("DB_PASSWORD", "socialfi_pg_pass_2024"),
			DBName:          getEnv("DB_NAME", "socialfi"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxConnections:  getEnvAsInt("DB_MAX_CONNECTIONS", 100),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 20),
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
			DBName:   getEnv("MYSQL_DB", "easitradecoins"),
		},
		Redis: RedisConfig{
			Host:       getEnv("REDIS_HOST", "localhost"),
			Port:       getEnv("REDIS_PORT", "6379"),
			Password:   getEnv("REDIS_PASSWORD", ""),
			DB:         getEnvAsInt("REDIS_DB", 0),
			MaxRetries: getEnvAsInt("REDIS_MAX_RETRIES", 3),
			PoolSize:   getEnvAsInt("REDIS_POOL_SIZE", 50),
		},
		Kafka: KafkaConfig{
			Brokers:            []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			TopicOrders:        getEnv("KAFKA_TOPIC_ORDERS", "orders"),
			TopicTrades:        getEnv("KAFKA_TOPIC_TRADES", "trades"),
			TopicNotifications: getEnv("KAFKA_TOPIC_NOTIFICATIONS", "notifications"),
			TopicRiskEvents:    getEnv("KAFKA_TOPIC_RISK_EVENTS", "risk_events"),
			TopicMarketData:    getEnv("KAFKA_TOPIC_MARKET_DATA", "market_data"),
			GroupID:            getEnv("KAFKA_GROUP_ID", "easitrade-consumer-group"),
		},
		Elasticsearch: ElasticsearchConfig{
			Host:   getEnv("ELASTICSEARCH_HOST", "localhost"),
			Port:   getEnv("ELASTICSEARCH_PORT", "9200"),
			Scheme: getEnv("ELASTICSEARCH_SCHEME", "http"),
		},
		Security: SecurityConfig{
			JWTSecret:     getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			JWTExpiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
			BCryptCost:    getEnvAsInt("BCRYPT_COST", 10),
		},
		Blockchain: BlockchainConfig{
			RPCURL:                      getEnv("ETHEREUM_RPC_URL", "https://sepolia.infura.io/v3/YOUR-PROJECT-ID"),
			ChainID:                     getEnvAsInt64("ETHEREUM_CHAIN_ID", 11155111),
			PrivateKey:                  getEnv("PRIVATE_KEY", ""),
			ContractAddressStaking:      getEnv("CONTRACT_ADDRESS_STAKING", ""),
			ContractAddressAirdrop:      getEnv("CONTRACT_ADDRESS_AIRDROP", ""),
			ContractAddressTokenFactory: getEnv("CONTRACT_ADDRESS_TOKEN_FACTORY", ""),
			ContractAddressDEXAggregator: getEnv("CONTRACT_ADDRESS_DEX_AGGREGATOR", ""),
			ContractAddressLiquidityMining: getEnv("CONTRACT_ADDRESS_LIQUIDITY_MINING", ""),
		},
		Risk: RiskConfig{
			OrderRateLimit:           getEnvAsInt("ORDER_RATE_LIMIT", 10),
			OrderRateWindow:          getEnvAsDuration("ORDER_RATE_WINDOW", 60*time.Second),
			MaxPriceDeviation:        getEnvAsFloat64("MAX_PRICE_DEVIATION", 0.20),
			WithdrawalDailyLimit:     getEnvAsFloat64("WITHDRAWAL_DAILY_LIMIT", 100000),
			LargeWithdrawalThreshold: getEnvAsFloat64("LARGE_WITHDRAWAL_THRESHOLD", 10000),
		},
		Features: FeatureFlags{
			EnableWebSocket:        getEnvAsBool("ENABLE_WEBSOCKET", true),
			EnableStopOrderMonitor: getEnvAsBool("ENABLE_STOP_ORDER_MONITOR", true),
			EnableRiskManager:      getEnvAsBool("ENABLE_RISK_MANAGER", true),
			EnableSwagger:          getEnvAsBool("ENABLE_SWAGGER", true),
		},
		Monitoring: MonitoringConfig{
			PrometheusEnabled: getEnvAsBool("PROMETHEUS_ENABLED", true),
			LogLevel:          getEnv("LOG_LEVEL", "info"),
		},
	}

	return cfg, nil
}

// GetDSN returns PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetMySQLDSN returns MySQL connection string
func (c *MySQLConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

// GetRedisAddr returns Redis address
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// GetElasticsearchURL returns Elasticsearch URL
func (c *ElasticsearchConfig) GetURL() string {
	return fmt.Sprintf("%s://%s:%s", c.Scheme, c.Host, c.Port)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat64(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
