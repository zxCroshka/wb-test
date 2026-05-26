package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/zxCroshka/wb-test/internal/aggregator"
	"github.com/zxCroshka/wb-test/internal/app/httpapp"
	"github.com/zxCroshka/wb-test/internal/cache"
	"github.com/zxCroshka/wb-test/internal/consumer"
	"github.com/zxCroshka/wb-test/internal/detector"
	"github.com/zxCroshka/wb-test/internal/domain"
	"github.com/zxCroshka/wb-test/internal/metrics"
	"github.com/zxCroshka/wb-test/internal/stoplist"
)

type App struct {
	HttpSrv    *httpapp.App
	Consumer   *consumer.Consumer
	Aggregator *aggregator.Aggregator
}

func New(
	ctx context.Context,
	log *slog.Logger,
	port int,
	bucketSize time.Duration,
	windowSize time.Duration,
	rps int,
	kafkaBrokers []string,
	kafkaTopic string,
	kafkaGroupID string,
	filepath string,
	rt time.Duration,
	wt time.Duration,
	it time.Duration,

) *App {
	window := domain.NewSlidingWindow(bucketSize, windowSize)
	cache := cache.NewTopCache()
	stopList, err := stoplist.NewStopListFromFile(filepath)
	if err != nil {
		log.Error("failed to create stopList from file", slog.String("error", err.Error()))
		panic(err)
	}
	detector := detector.NewAnomalyDetector(rps)
	metrics := metrics.NewMetrics()
	aggregator := aggregator.NewAggregator(window, cache, stopList, detector, metrics, log, filepath)
	aggregator.StartUpdater(ctx, 1*time.Second)
	httpserver := httpapp.New(log, aggregator, metrics, port, rt, wt, it)
	kafkaConfig := consumer.ConsumerConfig{
		Brokers: kafkaBrokers,
		Topic:   kafkaTopic,
		GroupID: kafkaGroupID,
	}
	Consumer, err := consumer.NewConsumer(kafkaConfig, aggregator, log)
	if err != nil {
		log.Error("failed to create kafka consumer", slog.String("error", err.Error()))
	}
	return &App{
		HttpSrv:    httpserver,
		Consumer:   Consumer,
		Aggregator: aggregator,
	}
}

func (a *App) SaveStopList() error {
	return a.Aggregator.SaveStopList()
}
