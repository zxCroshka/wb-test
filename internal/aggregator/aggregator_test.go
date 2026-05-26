package aggregator

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/zxCroshka/wb-test/internal/cache"
	"github.com/zxCroshka/wb-test/internal/detector"
	"github.com/zxCroshka/wb-test/internal/domain"
	"github.com/zxCroshka/wb-test/internal/stoplist"
)

var ctx = context.Background()

func NewTestAggregator(
	window *domain.SlidingWindow,
	topCache Cache,
	stopList StopList,
	detector AnomalyDetector,
) *Aggregator {
	logger := slog.Default() // используем дефолтный логер для тестов
	return NewAggregator(window, topCache, stopList, detector, nil, logger, "")
}

func TestAggregator_ProcessEvent(t *testing.T) {
	window := domain.NewSlidingWindow(1*time.Second, 5*time.Minute)
	topCache := cache.NewTopCache()
	stopList := stoplist.NewStopList()
	anomalyDetector := detector.NewAnomalyDetector(100)

	agg := NewTestAggregator(window, topCache, stopList, anomalyDetector)

	agg.ProcessEvent(ctx, "iphone", "user1")
	agg.ProcessEvent(ctx, "iphone", "user1")

	stats := window.GetStats()
	if stats["iphone"] != 2 {
		t.Errorf("Expected 2, got %d", stats["iphone"])
	}
}

func TestAggregator_StopListBlock(t *testing.T) {
	window := domain.NewSlidingWindow(1*time.Second, 5*time.Minute)
	topCache := cache.NewTopCache()
	stopList := stoplist.NewStopList()
	anomalyDetector := detector.NewAnomalyDetector(100)

	stopList.Add("spam")
	agg := NewTestAggregator(window, topCache, stopList, anomalyDetector)

	agg.ProcessEvent(ctx, "spam", "user1")

	stats := window.GetStats()
	if _, exists := stats["spam"]; exists {
		t.Error("Blocked word should not be in window")
	}
}

func TestAggregator_AnomalyDetectorBlock(t *testing.T) {
	window := domain.NewSlidingWindow(1*time.Second, 5*time.Minute)
	topCache := cache.NewTopCache()
	stopList := stoplist.NewStopList()
	anomalyDetector := detector.NewAnomalyDetector(1) // 1 запрос в секунду

	agg := NewTestAggregator(window, topCache, stopList, anomalyDetector)

	agg.ProcessEvent(ctx, "iphone", "user1")
	agg.ProcessEvent(ctx, "iphone", "user1")

	stats := window.GetStats()
	if stats["iphone"] != 1 {
		t.Errorf("Expected 1, got %d (anomaly should be blocked)", stats["iphone"])
	}
}

func TestAggregator_DifferentUsers(t *testing.T) {
	window := domain.NewSlidingWindow(1*time.Second, 5*time.Minute)
	topCache := cache.NewTopCache()
	stopList := stoplist.NewStopList()
	anomalyDetector := detector.NewAnomalyDetector(2)

	agg := NewTestAggregator(window, topCache, stopList, anomalyDetector)

	agg.ProcessEvent(ctx, "iphone", "user1")
	agg.ProcessEvent(ctx, "iphone", "user1")

	agg.ProcessEvent(ctx, "iphone", "user2")

	stats := window.GetStats()
	if stats["iphone"] != 3 {
		t.Errorf("Expected 3, got %d", stats["iphone"])
	}
}

func TestAggregator_GetTop(t *testing.T) {
	window := domain.NewSlidingWindow(1*time.Second, 5*time.Minute)
	topCache := cache.NewTopCache()
	stopList := stoplist.NewStopList()
	anomalyDetector := detector.NewAnomalyDetector(100)

	agg := NewTestAggregator(window, topCache, stopList, anomalyDetector)

	for i := 0; i < 10; i++ {
		agg.ProcessEvent(ctx, "iphone", "user1")
	}
	for i := 0; i < 5; i++ {
		agg.ProcessEvent(ctx, "samsung", "user1")
	}
	for i := 0; i < 3; i++ {
		agg.ProcessEvent(ctx, "xiaomi", "user1")
	}

	agg.updateTop()

	top, err := agg.GetTop(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(top) != 2 {
		t.Errorf("Expected 2 items, got %d", len(top))
	}
	if top[0].Query != "iphone" || top[0].Count != 10 {
		t.Errorf("Expected iphone=10, got %+v", top[0])
	}
	if top[1].Query != "samsung" || top[1].Count != 5 {
		t.Errorf("Expected samsung=5, got %+v", top[1])
	}

}
