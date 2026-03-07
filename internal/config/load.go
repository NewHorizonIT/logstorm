package config

import (
	"strings"

	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file (ignore error if file doesn't exist)
	_ = viper.ReadInConfig()

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
		},
		Kafka: KafkaConfig{
			Brokers:       parseStringSlice(viper.GetString("KAFKA_BROKERS")),
			ClientID:      viper.GetString("KAFKA_CLIENT_ID"),
			ConsumerGroup: viper.GetString("KAFKA_CONSUMER_GROUP"),
			Topics:        parseStringSlice(viper.GetString("KAFKA_TOPICS")),
			Linger:        viper.GetDuration("KAFKA_LINGER"),
			BatchSize:     viper.GetInt32("KAFKA_BATCH_SIZE"),
			RetryTimeout:  viper.GetDuration("KAFKA_RETRY_TIMEOUT"),
		},
		ClickHouse: ClickHouseConfig{
			Addr:        viper.GetString("CLICKHOUSE_ADDR"),
			Database:    viper.GetString("CLICKHOUSE_DATABASE"),
			Username:    viper.GetString("CLICKHOUSE_USERNAME"),
			Password:    viper.GetString("CLICKHOUSE_PASSWORD"),
			DialTimeout: viper.GetDuration("CLICKHOUSE_DIAL_TIMEOUT"),
		},
		Retry: RetryConfig{
			MaxRetries: viper.GetInt("RETRY_MAX_RETRIES"),
			BaseDelay:  viper.GetDuration("RETRY_BASE_DELAY"),
			MaxDelay:   viper.GetDuration("RETRY_MAX_DELAY"),
		},
	}

	return cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("SERVER_PORT", "3123")

	// Kafka defaults
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")
	viper.SetDefault("KAFKA_CLIENT_ID", "logstorm")
	viper.SetDefault("KAFKA_CONSUMER_GROUP", "logstorm-processor-group")
	viper.SetDefault("KAFKA_TOPICS", "logs-topic,dlq-log")
	viper.SetDefault("KAFKA_LINGER", "100ms")
	viper.SetDefault("KAFKA_BATCH_SIZE", 30000000)
	viper.SetDefault("KAFKA_RETRY_TIMEOUT", "30s")

	// ClickHouse defaults
	viper.SetDefault("CLICKHOUSE_ADDR", "localhost:9000")
	viper.SetDefault("CLICKHOUSE_DATABASE", "logstorm")
	viper.SetDefault("CLICKHOUSE_USERNAME", "logstorm")
	viper.SetDefault("CLICKHOUSE_PASSWORD", "123456")
	viper.SetDefault("CLICKHOUSE_DIAL_TIMEOUT", "5s")

	// Retry defaults
	viper.SetDefault("RETRY_MAX_RETRIES", 3)
	viper.SetDefault("RETRY_BASE_DELAY", "100ms")
	viper.SetDefault("RETRY_MAX_DELAY", "5s")
}

func parseStringSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
