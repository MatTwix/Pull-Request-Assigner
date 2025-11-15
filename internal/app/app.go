package app

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/MatTwix/Pull-Request-Assigner/internal/config"
	v1 "github.com/MatTwix/Pull-Request-Assigner/internal/controller/http/v1"
	"github.com/MatTwix/Pull-Request-Assigner/internal/repo"
	"github.com/MatTwix/Pull-Request-Assigner/internal/service"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/database/postgres"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/httpserver"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/logger"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// @title           Pull Request Assigner Service
// @version         1.0
// @description     Сервис для автоматического назначения ревьюверов на Pull Request'ы, управления командами и пользователями.

// @contact.name   Матвей Федоров
// @contact.tg @mattwix
// @contact.emal mfgolden@yandex.ru

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        X-Api-Key
// @description                 API key required for accessing protected endpoints
func Run(configPath string) {
	// Configuration
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		logrus.WithField("error", err).Fatal("config error")
	}

	// Logger
	log := logger.NewLogger(cfg.Env)

	// Database connection
	log.Info("initializing postgres...")
	pg, err := postgres.New(cfg.Postgres.Url)
	if err != nil {
		log.Fatal("failed to make new postgres connection, quitting service...", map[string]any{"error": err})
	}
	defer pg.Close()

	// Running migrations
	if err := runMigrations(log, cfg.Postgres.Url); err != nil {
		log.Fatal("failed to run migrations", map[string]any{"error": err})
	}

	// Repositories
	log.Info("initializing repositories...")
	repositories := repo.NewRepositories(pg)

	// Services dependencies
	log.Info("initializing services...")
	deps := service.ServicesDependencies{
		Repos:       repositories,
		AdminAPIKey: cfg.Auth.AdminAPIKey,
		UserAPIKey:  cfg.Auth.UserAPIKey,
	}
	services := service.NewServices(deps)

	// Handlers and routes
	log.Info("initializing handlers and routes...")
	handler := chi.NewRouter()
	// initing handlers validator
	utils.InitValidator()
	v1.NewRouter(handler, services, log)

	// HTTP server
	log.Info("starting http server...")
	log.Debug("info", map[string]any{"port": cfg.HttpServer.Port})
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HttpServer.Port))

	// Waiting signal
	log.Info("configuring graceful shutdown...")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("catched interrupt signal", map[string]any{"signal": s})
	case err = <-httpServer.Notify():
		log.Error("catched http server error signal", map[string]any{"signal": err})
	}

	// Graceful shutdown
	log.Info("Shutting down...")
	if err = httpServer.Shutdown(); err != nil {
		log.Error("failed to shut down http server", map[string]any{"error": err})
	}
}
