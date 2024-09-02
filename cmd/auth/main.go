package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DarRo9/Test-task-BackDev/internal/auth"
	"github.com/DarRo9/Test-task-BackDev/internal/config"
	"github.com/DarRo9/Test-task-BackDev/internal/http-server/handler"
	"github.com/DarRo9/Test-task-BackDev/internal/http-server/middleware/logger"
	"github.com/DarRo9/Test-task-BackDev/internal/lib/logger/sl"
	"github.com/DarRo9/Test-task-BackDev/internal/server"
	"github.com/DarRo9/Test-task-BackDev/internal/service"
	"github.com/DarRo9/Test-task-BackDev/internal/storage/mongodb"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

const (
	dev   = "dev"
	prod  = "prod"
	local = "local"
)

func main() {
	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatal("Failed loading .env file")
	}

	config := config.Loading()

	log := setupLogger(config.Env)

	log.Info(
		"Starting",
		slog.String("env", config.Env))
	log.Debug("Messages about debug are enabled")

	mongoClient, err := mongodb.CreateObjectClient(config.Mongo.URI, config.Mongo.User, config.Mongo.Password)
	if err != nil {
		log.Error("Fail of initiation storage", sl.Err(err))
		os.Exit(1)
	}

	mongoDatabase := mongodb.CreateObjectStorage(mongoClient, config.Mongo.Database)
	mongoRefreshRepo := mongoDatabase.CreateObjectRefreshRepo()

	tokenAuthenticator, err := auth.CreateObject(config.JWT.SigningKey)
	if err != nil {
		log.Error("Fail of initiation auth", sl.Err(err))
		os.Exit(1)
	}

	service, err := service.CreateObject(config, mongoRefreshRepo, tokenAuthenticator)
	if err != nil {
		log.Error("Fail of initiation service", sl.Err(err))
		os.Exit(1)
	}

	logger := logger.Log

	h := handler.CreateObject(config, service, logger)

	srv := server.CreateObject(config, h.CreateObjectRouter())

	go func() {
		if err := srv.Start(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("Fail of initiation http server", sl.Err(err))
			os.Exit(1)
		}
	}()

	log.Info("Start server")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		log.Error("Fail of stopping server", sl.Err(err))
		os.Exit(1)
	}

	if err := mongoClient.Disconnect(context.Background()); err != nil {
		log.Error("Fail of stopping mongo client", sl.Err(err))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case local:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case dev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case prod:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
