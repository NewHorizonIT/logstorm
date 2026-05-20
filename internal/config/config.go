package config

import (
	"time"
)

type Config struct {
	Server     ServerConfig
	Kafka      KafkaConfig
	ClickHouse ClickHouseConfig
	Retry      RetryConfig
	Database   DatabaseConfig
}

type ServerConfig struct {
	Port string `mapstructure:"SERVER_PORT"`
}

type KafkaConfig struct {
	Brokers       []string      `mapstructure:"KAFKA_BROKERS"`
	ClientID      string        `mapstructure:"KAFKA_CLIENT_ID"`
	ConsumerGroup string        `mapstructure:"KAFKA_CONSUMER_GROUP"`
	Topics        []string      `mapstructure:"KAFKA_TOPICS"`
	Linger        time.Duration `mapstructure:"KAFKA_LINGER"`
	BatchSize     int32         `mapstructure:"KAFKA_BATCH_SIZE"`
	RetryTimeout  time.Duration `mapstructure:"KAFKA_RETRY_TIMEOUT"`
}

type ClickHouseConfig struct {
	Addr        string        `mapstructure:"CLICKHOUSE_ADDR"`
	Database    string        `mapstructure:"CLICKHOUSE_DATABASE"`
	Username    string        `mapstructure:"CLICKHOUSE_USERNAME"`
	Password    string        `mapstructure:"CLICKHOUSE_PASSWORD"`
	DialTimeout time.Duration `mapstructure:"CLICKHOUSE_DIAL_TIMEOUT"`
}

type RetryConfig struct {
	MaxRetries int           `mapstructure:"RETRY_MAX_RETRIES"`
	BaseDelay  time.Duration `mapstructure:"RETRY_BASE_DELAY"`
	MaxDelay   time.Duration `mapstructure:"RETRY_MAX_DELAY"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	Port     string `mapstructure:"DB_PORT"`
	SSLMode  string `mapstructure:"DB_SSL_MODE"`
}
