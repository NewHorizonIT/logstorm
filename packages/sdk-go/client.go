package sdkgo

import (
	"context"
	"sync"
)

// Client sends log events to the LogStorm ingestion API.
// Create a Client with New, use it to emit log events, and call Shutdown
// before your program exits to flush any buffered events.
//
//	client, err := sdkgo.New("lsk_xxx", sdkgo.WithService("my-service"))
//	if err != nil { ... }
//	defer client.Shutdown(context.Background())
//
//	client.Info("user logged in", sdkgo.F("user_id", "123"))
type Client struct {
	cfg    Config
	ch     chan *logEvent
	done   chan struct{}
	wg     sync.WaitGroup
	sender *httpSender
}

// New creates a Client with the given API key and starts the background worker.
// Defaults: endpoint=https://api.logstorm.io, batchSize=50, flushInterval=500ms, channelSize=1000.
// Returns an error if apiKey is empty or any option value is out of range.
func New(apiKey string, opts ...Option) (*Client, error) {
	cfg := Config{
		Endpoint:      defaultEndpoint,
		APIKey:        apiKey,
		BatchSize:     defaultBatchSize,
		FlushInterval: defaultFlushInterval,
		ChannelSize:   defaultChannelSize,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	c := &Client{
		cfg:    cfg,
		ch:     make(chan *logEvent, cfg.ChannelSize),
		done:   make(chan struct{}),
		sender: newHTTPSender(cfg.Endpoint, cfg.APIKey),
	}

	c.wg.Add(1)
	go c.worker()

	return c, nil
}

// Shutdown drains all buffered events, sends them to the API, and stops the
// background worker. It blocks until the flush completes or ctx is cancelled.
// Always call Shutdown before your program exits to avoid losing buffered events.
func (c *Client) Shutdown(ctx context.Context) error {
	close(c.done)

	finished := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(finished)
	}()

	select {
	case <-finished:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) enqueue(e *logEvent) {
	select {
	case c.ch <- e:
	default:
		// buffer full — drop to never block the caller
	}
}
