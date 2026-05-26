package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zxCroshka/wb-test/configs"
	"github.com/zxCroshka/wb-test/internal/app"
	"github.com/zxCroshka/wb-test/internal/logger"
)

var (
	configPath = flag.String("config", "configs/config.yaml", "path to config file")
)

func main() {
	flag.Parse()

	cfg, err := configs.Load(*configPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(logger.Config{
		Level:     cfg.Logging.Level,
		Format:    cfg.Logging.Format,
		AddSource: cfg.Logging.AddSource,
	})
	log.Info("starting trending search service", slog.String("config", *configPath))
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	application := app.New(
		shutdownCtx,
		log,
		cfg.HTTP.Port,
		cfg.Business.BucketSize,
		cfg.Business.WindowSize,
		cfg.AnomalyDetector.RequestsPerSecond,
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		cfg.Kafka.GroupID,
		cfg.Stoplist.PersistFile,
		cfg.HTTP.ReadTimeout,
		cfg.HTTP.WriteTimeout,
		cfg.HTTP.IdleTimeout,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if application.Consumer != nil {
		if err := application.Consumer.Start(ctx); err != nil {
			log.Error("failed to start kafka consumer", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	go func() {
		log.Info("starting HTTP server", slog.String("port", "8080"))
		if err := application.HttpSrv.Run(); err != nil {
			log.Error("HTTP server error", slog.String("error", err.Error()))
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gracefully...")
	application.SaveStopList()
	if application.Consumer != nil {
		application.Consumer.Stop()
	}

	application.Aggregator.Stop()

	if err := application.HttpSrv.Stop(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown error", slog.String("error", err.Error()))
	}

	cancel()
	log.Info("service stopped successfully")
}
