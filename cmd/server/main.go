package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"topikkhusus-methodtracker/internal/config"
	"topikkhusus-methodtracker/internal/handlers"
	"topikkhusus-methodtracker/internal/middleware"
	"topikkhusus-methodtracker/internal/repository"
	"topikkhusus-methodtracker/internal/services"
	"topikkhusus-methodtracker/internal/tracker"
	redisclient "topikkhusus-methodtracker/pkg/redis"
	"topikkhusus-methodtracker/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	client, err := redisclient.NewClient(cfg)
	if err != nil {
		log.Fatalf("init redis client failed: %v", err)
	}

	repositoryLayer := repository.NewRedisUserRepository(client, cfg.RedisTimeout)
	serviceLayer := services.NewUserService(repositoryLayer)
	trackerLayer := tracker.NewMethodTracker(client, cfg.RedisTimeout)

	healthCheck := func(ctx context.Context) error {
		return redisclient.HealthCheck(ctx, client, cfg.RedisTimeout)
	}

	handlerLayer := handlers.NewUserHandler(serviceLayer, healthCheck)

	httpHandler := routes.Register(
		handlerLayer,
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Tracker(trackerLayer.Track),
		middleware.Recovery(),
	)

	secureHandler := middleware.Chain(
		httpHandler,
		middleware.RateLimit(client, cfg.RateLimitRPM, cfg.RedisTimeout),
	)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.AppPort),
		Handler:           secureHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("server started on port %s", cfg.AppPort)
		if serveErr := server.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", serveErr)
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-shutdownCtx.Done()
	log.Printf("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	if err = client.Close(); err != nil {
		log.Printf("redis close failed: %v", err)
	}

	log.Printf("server exited gracefully")
	os.Exit(0)
}
