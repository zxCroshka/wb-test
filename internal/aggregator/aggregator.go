package aggregator

import (
	"context"
	"log/slog"
	"time"

	"github.com/zxCroshka/wb-test/internal/domain"
)

type MetricsCollector interface {
	IncSearchEvents()
	IncBlockedQueries()
	IncAnomalyQueries()
	IncCacheHits()
	IncCacheMisses()
	SetTopCacheSize(size int)
	SetWindowStats(totalQueries, uniqueQueries int64, activeBuckets int)
	ObserveAggregationDuration(duration float64)
}

type StopList interface {
	Add(string)
	Remove(string) bool
	Contains(string) bool
	GetAll() []string
	Size() int
	BatchAdd([]string) int
	Clear()
	SaveToFile(filePath string) error
}

type AnomalyDetector interface {
	IsAnomaly(query string, userID string) bool
}

type Cache interface {
	GetTop(limit int) ([]domain.RankItem, error)
	UpdateFromStats(stats map[string]int64, maxSize int)
	Size() int
}

type Aggregator struct {
	window           *domain.SlidingWindow
	topCache         Cache
	stopList         StopList
	detector         AnomalyDetector
	metrics          MetricsCollector
	logger           *slog.Logger
	stopCh           chan struct{}
	updaterRunning   bool
	stopListFilePath string
}

func NewAggregator(
	window *domain.SlidingWindow,
	topCache Cache,
	stopList StopList,
	detector AnomalyDetector,
	metrics MetricsCollector,
	logger *slog.Logger,
	filepath string,
) *Aggregator {
	return &Aggregator{
		window:           window,
		topCache:         topCache,
		stopList:         stopList,
		detector:         detector,
		metrics:          metrics,
		logger:           logger.WithGroup("aggregator"),
		stopCh:           make(chan struct{}),
		updaterRunning:   false,
		stopListFilePath: filepath,
	}
}

func (a *Aggregator) ProcessEvent(ctx context.Context, query, userID string) {
	select {
	case <-ctx.Done():
		a.logger.Warn("event processing cancelled",
			slog.String("query", query),
			slog.String("user_id", userID),
			slog.String("reason", ctx.Err().Error()),
		)
		return
	default:
	}

	defer func() {
		if r := recover(); r != nil {
			a.logger.Error("panic in ProcessEvent",
				slog.Any("recover", r),
				slog.String("query", query),
				slog.String("user_id", userID),
			)
		}
	}()

	if a.metrics != nil {
		a.metrics.IncSearchEvents()
	}

	if a.stopList != nil && a.stopList.Contains(query) {
		if a.metrics != nil {
			a.metrics.IncBlockedQueries()
		}
		a.logger.Debug("query blocked by stoplist",
			slog.String("query", query),
			slog.String("user_id", userID),
		)
		return
	}

	if a.detector != nil && a.detector.IsAnomaly(query, userID) {
		if a.metrics != nil {
			a.metrics.IncAnomalyQueries()
		}
		a.logger.Warn("anomaly detected",
			slog.String("query", query),
			slog.String("user_id", userID),
		)
		return
	}

	a.window.Add(query)

	a.logger.Debug("event processed",
		slog.String("query", query),
		slog.String("user_id", userID),
	)
}

func (a *Aggregator) GetTop(ctx context.Context, limit int) ([]domain.RankItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	top, err := a.topCache.GetTop(limit)

	if a.metrics != nil {
		if err == nil && len(top) > 0 {
			a.metrics.IncCacheHits()
			a.logger.Debug("cache hit", slog.Int("limit", limit), slog.Int("result_count", len(top)))
		} else {
			a.metrics.IncCacheMisses()
			a.logger.Debug("cache miss", slog.Int("limit", limit))
		}
	}

	if err != nil {
		a.logger.Error("failed to get top",
			slog.Int("limit", limit),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return top, nil
}

func (a *Aggregator) AddToStopList(word string) {
	if a.stopList != nil {
		a.stopList.Add(word)
		a.logger.Info("word added to stoplist", slog.String("word", word))
	}
}

func (a *Aggregator) RemoveFromStopList(word string) bool {
	if a.stopList == nil {
		return false
	}
	removed := a.stopList.Remove(word)
	if removed {
		a.logger.Info("word removed from stoplist", slog.String("word", word))
	} else {
		a.logger.Warn("word not found in stoplist", slog.String("word", word))
	}
	return removed
}

func (a *Aggregator) IsWordBlocked(word string) bool {
	if a.stopList == nil {
		return false
	}
	return a.stopList.Contains(word)
}

func (a *Aggregator) GetStopList() []string {
	if a.stopList == nil {
		return []string{}
	}
	return a.stopList.GetAll()
}

func (a *Aggregator) GetStopListSize() int {
	if a.stopList == nil {
		return 0
	}
	return a.stopList.Size()
}

func (a *Aggregator) BatchAddToStopList(words []string) int {
	if a.stopList == nil {
		return 0
	}
	added := a.stopList.BatchAdd(words)
	a.logger.Info("batch added to stoplist",
		slog.Int("added", added),
		slog.Int("total", len(words)),
	)
	return added
}

func (a *Aggregator) ClearStopList() {
	if a.stopList != nil {
		a.stopList.Clear()
		a.logger.Warn("stoplist cleared")
	}
}

func (a *Aggregator) updateTop() {
	start := time.Now()

	defer func() {
		if r := recover(); r != nil {
			a.logger.Error("panic in updateTop", slog.Any("recover", r))
		}
	}()

	stats := a.window.GetStats()

	a.topCache.UpdateFromStats(stats, 100)

	if a.metrics != nil {
		duration := time.Since(start).Seconds()
		a.metrics.ObserveAggregationDuration(duration)
		a.metrics.SetTopCacheSize(a.topCache.Size())

		totalQueries, uniqueQueries, activeBuckets := a.GetWindowStats()
		a.metrics.SetWindowStats(totalQueries, uniqueQueries, activeBuckets)

		a.logger.Debug("top updated",
			slog.Int("unique_queries", len(stats)),
			slog.Int("cache_size", a.topCache.Size()),
			slog.Int64("total_queries", totalQueries),
			slog.Int("active_buckets", activeBuckets),
			slog.Duration("duration", time.Since(start)),
		)
	}
}

func (a *Aggregator) GetWindowStats() (totalQueries, uniqueQueries int64, activeBuckets int) {
	stats := a.window.GetStats()

	for _, count := range stats {
		totalQueries += count
	}
	uniqueQueries = int64(len(stats))
	activeBuckets = a.window.GetActiveBucketsCount()

	return
}

func (a *Aggregator) StartUpdater(ctx context.Context, interval time.Duration) {
	if a.updaterRunning {
		a.logger.Info("updater already running")
		return
	}
	a.logger.Info("starting top updater", slog.Duration("interval", interval))

	ticker := time.NewTicker(interval)
	go func() {
		defer func() {
			ticker.Stop()
			a.updaterRunning = false
			a.logger.Info("top updater stopped")
		}()
		for {
			select {
			case <-ticker.C:
				a.updateTop()
			case <-a.stopCh:
				a.logger.Info("top updater stopped")
				return
			case <-ctx.Done():
				a.logger.Info("updater stopped by context", slog.String("reason", ctx.Err().Error()))
				return
			}
		}
	}()
}

func (a *Aggregator) Stop() {
	a.logger.Info("stopping aggregator")
	close(a.stopCh)

	a.logger.Info("aggregator stopped")
}

func (a *Aggregator) SaveStopList() error {
	if a.stopList == nil {
		return nil
	}
	return a.stopList.SaveToFile(a.stopListFilePath)
}
