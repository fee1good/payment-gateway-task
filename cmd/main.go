package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"payment-gateway/configs/envs"
	"payment-gateway/configs/logger"
	"payment-gateway/db"
	"payment-gateway/internal/api"
	"payment-gateway/internal/cache"
	"payment-gateway/internal/gateway"
	"payment-gateway/internal/kafka"
	"payment-gateway/internal/services"
	"payment-gateway/internal/workers"
)

func main() {
	cfg := envs.Load()

	logger.Init(cfg.LogLevel)
	logger.Info("Starting payment gateway service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.InitDB(cfg.DB.URL)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	redisClient := cache.InitRedis(ctx, cfg.Redis.Addr, cfg.Redis.Password)

	var kafkaProducer kafka.Producer
	if len(cfg.Kafka.Brokers) > 0 {
		kafkaProducer = kafka.NewProducer(cfg.Kafka.Brokers[0])
	} else {
		logger.Warn("No Kafka brokers configured, Kafka producer will not be initialized")
	}

	dbHandler := db.NewDBHandler(database)
	redisCache := cache.NewRedisCache(redisClient)
	stripeClient := gateway.Stripe(cfg)

	processor := workers.NewTransactionProcessor(dbHandler, cfg.Workers.Count, stripeClient)
	processor.Start(ctx)

	gatewayService := services.NewGateway(dbHandler, redisCache, processor, kafkaProducer, cfg)

	router := api.SetupRouter(dbHandler, gatewayService)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("Server starting", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// graceful shutdown
	<-stop
	logger.Info("Shutting down server...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	logger.Info("Shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server forced to shutdown", "error", err)
	}

	logger.Info("Stopping transaction processor...")
	processor.Stop()

	if kafkaProducer != nil {
		logger.Info("Closing Kafka producer...")
		if err := kafkaProducer.Close(); err != nil {
			logger.Error("Error closing Kafka producer", "error", err)
		}
	}

	logger.Info("Closing Redis connection...")
	if err := redisClient.Close(); err != nil {
		logger.Error("Error closing Redis connection", "error", err)
	}

	logger.Info("Closing database connection...")
	if err := database.Close(); err != nil {
		logger.Error("Error closing database connection", "error", err)
	}

	logger.Info("Server gracefully stopped")
}
