package sdkgo

import (
	"errors"
	"fmt"
	"time"
)

const (
	defaultEndpoint      = "https://api.logstorm.io"
	defaultBatchSize     = 50
	defaultFlushInterval = 500 * time.Millisecond
	defaultChannelSize   = 1000
)

// Config holds all configuration for a Client.
// Use Option functions with New rather than constructing Config directly.
type Config struct {
	Endpoint      string
	APIKey        string
	Service       string
	Environment   string
	BatchSize     int
	FlushInterval time.Duration
	ChannelSize   int
}

// Option is a function that configures a Client.
type Option func(*Config)

// WithEndpoint overrides the LogStorm API base URL.
// Defaults to https://api.logstorm.io.
func WithEndpoint(endpoint string) Option {
	return func(c *Config) { c.Endpoint = endpoint }
}

// WithService sets a default service name attached to every log event.
func WithService(service string) Option {
	return func(c *Config) { c.Service = service }
}

// WithEnvironment sets a default environment attached to every log event (e.g. "production", "staging").
func WithEnvironment(env string) Option {
	return func(c *Config) { c.Environment = env }
}

// WithBatchSize controls how many events are grouped into a single HTTP request.
// Valid range: 1–1000. Defaults to 50.
func WithBatchSize(n int) Option {
	return func(c *Config) { c.BatchSize = n }
}

// WithFlushInterval controls how often buffered events are sent even when the
// batch is not full. Valid range: 100ms–1h. Defaults to 500ms.
func WithFlushInterval(d time.Duration) Option {
	return func(c *Config) { c.FlushInterval = d }
}

// WithChannelSize sets the capacity of the internal event buffer.
// If the buffer is full, new events are dropped to avoid blocking the caller.
// Valid range: 1–100000. Defaults to 1000.
func WithChannelSize(n int) Option {
	return func(c *Config) { c.ChannelSize = n }
}

func (c *Config) validate() error {
	if c.APIKey == "" {
		return errors.New("api key is required")
	}
	if c.Endpoint == "" {
		return errors.New("endpoint must not be empty")
	}
	if c.BatchSize < 1 || c.BatchSize > 1000 {
		return fmt.Errorf("batch size must be between 1 and 1000, got %d", c.BatchSize)
	}
	if c.FlushInterval < 100*time.Millisecond || c.FlushInterval > time.Hour {
		return fmt.Errorf("flush interval must be between 100ms and 1h, got %s", c.FlushInterval)
	}
	if c.ChannelSize < 1 || c.ChannelSize > 100_000 {
		return fmt.Errorf("channel size must be between 1 and 100000, got %d", c.ChannelSize)
	}
	return nil
}
