package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/NewHorizonIT/logstorm/internal/config"
	"github.com/NewHorizonIT/logstorm/internal/global"
	"github.com/NewHorizonIT/logstorm/internal/infra/clickhouse"
	"github.com/NewHorizonIT/logstorm/internal/infra/kafka"
	"github.com/NewHorizonIT/logstorm/internal/infra/postgres"
	"github.com/NewHorizonIT/logstorm/internal/infra/redis"
	"github.com/NewHorizonIT/logstorm/internal/observability"
	"github.com/NewHorizonIT/logstorm/internal/services/auth"
	"github.com/NewHorizonIT/logstorm/internal/services/ingestion"
	"github.com/NewHorizonIT/logstorm/internal/services/processor"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	global.GlobalConfig = *cfg

	slog.Info("configuration loaded", "config", cfg)

	// Context and signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup Kafka (shared client for producer & consumer)
	kafkaConfig := kafka.Config{
		Brokers:       cfg.Kafka.Brokers,
		ClientID:      cfg.Kafka.ClientID,
		ConsumerGroup: cfg.Kafka.ConsumerGroup,
		Topics:        cfg.Kafka.Topics,
		Linger:        cfg.Kafka.Linger,
		BatchSize:     cfg.Kafka.BatchSize,
		RetryTimeout:  cfg.Kafka.RetryTimeout,
	}

	kafkaClient, err := kafka.InitKafkaClient(kafkaConfig)
	if err != nil {
		panic(err)
	}
	defer kafkaClient.Close()

	// Producer & Consumer share same client
	publisher := kafka.NewKafkaProducer(kafkaClient)
	consumer := kafka.NewKafkaConsumer(kafkaClient)

	// Setup ClickHouse
	chConn, err := clickhouse.NewClickHouse(cfg.ClickHouse)
	if err != nil {
		panic(err)
	}
	defer chConn.Close()

	// Retry config for ClickHouse
	retryCnf := kafka.RetryCnf{
		MaxRetries: cfg.Retry.MaxRetries,
		BaseDelay:  cfg.Retry.BaseDelay,
		MaxDelay:   cfg.Retry.MaxDelay,
	}

	// Connect database
	db := postgres.Connect(&cfg.Database)

	slog.Info("[POSTGES]::Connected")

	// Initialize redis
	redisClient := redis.NewClient(cfg.Redis)
	defer redisClient.Close()

	slog.Info("[REDIS]::Connected")

	// Create redis cache
	redisCache := redis.NewRedisCache(redisClient)

	slog.Info("[REDIS CACHE]::Initialized")

	// Auth service
	if err := db.AutoMigrate(&auth.Account{}); err != nil {
		panic(err)
	}
	authRepo := auth.NewAuthRepository(db)
	sessionRepo := auth.NewSessionRepository(redisCache)
	authUsecase := auth.NewAuthUsecase(authRepo, sessionRepo)
	authHandler := auth.NewAuthHandler(authUsecase)
	authRouter := auth.NewAuthRouter(authHandler)

	// Repository
	repository := clickhouse.NewLogStormRepository(chConn, retryCnf, publisher)

	// Processor
	proc := processor.NewProcessor(consumer, repository, publisher)

	// Start processor in background
	go func() {
		slog.Info("starting log processor")
		if err := proc.Start(ctx); err != nil {
			slog.Error("processor error", "error", err)
		}
	}()

	// Setup Router
	router := gin.Default()

	// Add Prometheus middleware
	router.Use(observability.PrometheusMiddleware())

	// Metrics endpoint for Prometheus
	// router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth routes
	authGroup := router.Group("/auth")
	authRouter.RegisterRoutes(authGroup)

	// Ingestion routes
	ingestionGroup := router.Group("/ingestion")
	ingestion.SetupIngestionRoutes(ingestionGroup, publisher)

	// Define server
	serverAddr := ":" + cfg.Server.Port
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Shutdown server gracefully on interrupt signal
	go func() {
		slog.Info("starting HTTP server", "addr", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	// with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	slog.Info("shutting down...")
	cancel() // Cancel context to stop processor

	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()
	server.Shutdown(ctxTimeout)
}
