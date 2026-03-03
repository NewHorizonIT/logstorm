package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	handler "github.com/NewHorizonIT/logstorm-ingestion/internal/hanlder"
	"github.com/NewHorizonIT/logstorm-ingestion/internal/middleware"
	"github.com/NewHorizonIT/logstorm-ingestion/internal/producer"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create async producer
	prod := producer.NewKafkaProducer(
		"localhost:9092",
		"logs-topic",
		10000,
		4,
	)

	go prod.Start(ctx)

	router := gin.Default()

	router.Use(middleware.MetricsMiddleware())

	// Set up handler with producer and rate limiter
	limiter := middleware.NewManager(
		rate.Limit(10000),
		20000,
		5*time.Minute,
	)

	h := handler.NewIngestHandler(prod, limiter)

	router.POST("/ingest", h.HandleIngest)
	router.GET("/metrics", h.MetricsHandler())

	srv := &http.Server{
		Addr:    ":3123",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")

	cancel()
	prod.Close()

	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()
	srv.Shutdown(ctxTimeout)
}
