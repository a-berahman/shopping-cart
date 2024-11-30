package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/a-berahman/shopping-cart/config"
	"github.com/a-berahman/shopping-cart/internal/adapters/handler"
	"github.com/a-berahman/shopping-cart/internal/adapters/queue"
	"github.com/a-berahman/shopping-cart/internal/adapters/repository"
	"github.com/a-berahman/shopping-cart/internal/adapters/reservation"
	"github.com/a-berahman/shopping-cart/internal/adapters/reservation/mock"
	"github.com/a-berahman/shopping-cart/internal/core/ports"

	"github.com/a-berahman/shopping-cart/internal/service"
	"github.com/a-berahman/shopping-cart/internal/worker"
)

func main() {
	logger := slog.Default()

	// Load configuration
	configPath := flag.String("config-path", "config.yaml", "path to config file")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := initializeApplication(ctx, logger, cfg)
	if err != nil {
		logger.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}
	app.startServer(wg, cfg.Server.Host, cfg.Server.Port)

	<-ctx.Done()
	stop()
	logger.Info("shutting down application")

	if err := app.shutdown(ctx, wg); err != nil {
		logger.Error("failed to shut down gracefully", "error", err)
		os.Exit(1)
	}
}

type application struct {
	server *echo.Echo
	worker *worker.ReservationWorker
}

func initializeApplication(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*application, error) {
	// Database initialization
	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Redis queue initialization
	redisQueue, err := queue.NewRedisQueue(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis queue: %w", err)
	}

	// Repository setup
	repo := repository.NewRepository(db)

	// Reservation service setup
	var reservationSvc ports.ReservationService
	if cfg.Env == "production" {
		reservationSvc = reservation.NewService(cfg.Reservation.ServiceURL)
	} else {
		mockConfig := mock.MockConfig{
			LatencyRange: 2 * time.Second,
			FailureRate:  0.1,
		}
		reservationSvc = mock.NewMockReservationService(mockConfig)
	}

	// Service and worker setup
	cartService := service.NewCartService(repo, redisQueue, reservationSvc)
	worker := worker.NewReservationWorker(redisQueue, reservationSvc, repo)

	// Start worker
	go worker.Start(ctx)

	// Echo server setup
	server := setupEcho(logger)
	handler.NewHandler(cartService).Register(server)

	return &application{server: server, worker: worker}, nil
}

func (app *application) startServer(wg *sync.WaitGroup, host string, port int) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.server.Start(fmt.Sprintf("%s:%d", host, port)); err != nil {
			slog.Default().Error("server failed to start", "error", err)
		}
	}()
}

func (app *application) shutdown(ctx context.Context, wg *sync.WaitGroup) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := app.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	app.worker.Stop()
	wg.Wait()
	return nil
}

func setupEcho(logger *slog.Logger) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	// e.Use(requestTimer(logger))

	e.Validator = &CustomValidator{Validator: validator.New()}
	return e
}

// func requestTimer(logger *slog.Logger) echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			start := time.Now()
// 			err := next(c)
// 			duration := time.Since(start)

// 			logger.Info("request completed",
// 				"method", c.Request().Method,
// 				"path", c.Request().URL.Path,
// 				"status", c.Response().Status,
// 				"duration", duration.String(),
// 				"duration_ms", duration.Milliseconds(),
// 			)
// 			return err
// 		}
// 	}
// }

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}
