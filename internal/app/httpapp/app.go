package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/zxCroshka/wb-test/internal/aggregator"
	"github.com/zxCroshka/wb-test/internal/api"
	"github.com/zxCroshka/wb-test/internal/metrics"
)

type App struct {
	log    *slog.Logger
	server *api.Server
	port   int
	rt     time.Duration
	wt     time.Duration
	it     time.Duration
}

func New(
	log *slog.Logger,
	aggregator *aggregator.Aggregator,
	metrics *metrics.Metrics,
	port int,
	rt time.Duration,
	wt time.Duration,
	it time.Duration,
) *App {

	server := api.NewServer(aggregator, metrics) // ← передаём

	return &App{
		log:    log,
		server: server,
		port:   port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {

	addr := fmt.Sprintf(":%d", a.port)

	a.log.Info("Handler server is running", slog.String("addr", addr))

	if err := a.server.Start(addr, a.rt, a.wt, a.it); err != nil {
		return err
	}
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
