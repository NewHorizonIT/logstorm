package main

import (
	"context"
	"log"
	"time"

	sdkgo "github.com/NewHorizonIT/logstorm-sdk-go"
)

func main() {
	client, err := sdkgo.New("lsk_your_api_key_here",
		sdkgo.WithEndpoint("http://localhost:8080"),
		sdkgo.WithService("payment-service"),
		sdkgo.WithEnvironment("production"),
		sdkgo.WithBatchSize(50),
		sdkgo.WithFlushInterval(500*time.Millisecond),
	)
	if err != nil {
		log.Fatalf("logstorm: %v", err)
	}
	defer client.Shutdown(context.Background())

	// basic log levels
	client.Debug("cache miss", sdkgo.F("key", "user:123"))
	client.Info("user logged in", sdkgo.F("user_id", "123"), sdkgo.F("ip", "1.2.3.4"))
	client.Warn("rate limit approaching", sdkgo.F("user_id", "123"), sdkgo.F("requests", "950"))
	client.Error("payment failed", sdkgo.F("order_id", "ord_456"), sdkgo.F("reason", "insufficient funds"))
	client.Fatal("database unreachable", sdkgo.F("host", "db.internal"))

	// simulate some work so the batch has time to flush
	time.Sleep(time.Second)
}
