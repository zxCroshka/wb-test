package aggregator

import (
	"testing"
	"time"

	"github.com/zxCroshka/wb-test/internal/cache"
	"github.com/zxCroshka/wb-test/internal/detector"
	"github.com/zxCroshka/wb-test/internal/domain"
	"github.com/zxCroshka/wb-test/internal/stoplist"
)

// измеряет скорость обработки событий
func BenchmarkAggregator_ProcessEvent(b *testing.B) {
	window := domain.NewSlidingWindow(1*time.Second, 5*time.Minute)
	topCache := cache.NewTopCache()
	stopList := stoplist.NewStopList()
	detector := detector.NewAnomalyDetector(100)

	agg := NewTestAggregator(window, topCache, stopList, detector)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg.ProcessEvent(ctx, "iphone", "user1")
	}
}

// измеряет конкурентную обработку
func BenchmarkAggregator_ProcessEventParallel(b *testing.B) {
	window := domain.NewSlidingWindow(1*time.Second, 5*time.Minute)
	topCache := cache.NewTopCache()
	stopList := stoplist.NewStopList()
	detector := detector.NewAnomalyDetector(100)

	agg := NewTestAggregator(window, topCache, stopList, detector)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			agg.ProcessEvent(ctx, "iphone", "user1")
		}
	})
}

// измеряет скорость получения топа
func BenchmarkAggregator_GetTop(b *testing.B) {
	window := domain.NewSlidingWindow(1*time.Second, 5*time.Minute)
	topCache := cache.NewTopCache()
	stopList := stoplist.NewStopList()
	detector := detector.NewAnomalyDetector(100)

	agg := NewTestAggregator(window, topCache, stopList, detector)

	stats := make(map[string]int64)
	for i := 0; i < 100; i++ {
		stats["query"] = int64(i)
	}
	topCache.UpdateFromStats(stats, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg.GetTop(ctx, 10)
	}
}

func BenchmarkAggregator_Mixed(b *testing.B) {
	window := domain.NewSlidingWindow(1*time.Second, 5*time.Minute)
	topCache := cache.NewTopCache()
	stopList := stoplist.NewStopList()
	detector := detector.NewAnomalyDetector(100)

	agg := NewTestAggregator(window, topCache, stopList, detector)
	stats := make(map[string]int64)
	for i := 0; i < 100; i++ {
		stats["query"] = int64(i)
	}
	topCache.UpdateFromStats(stats, 100)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			agg.ProcessEvent(ctx, "iphone", "user1")
			agg.GetTop(ctx, 10)
		}
	})
}
