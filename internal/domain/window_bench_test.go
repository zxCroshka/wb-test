package domain

import (
	"testing"
	"time"
)

// измеряет скорость добавления запросов
func BenchmarkSlidingWindow_Add(b *testing.B) {
	window := NewSlidingWindow(1*time.Second, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		window.Add("iphone")
	}
}

// измеряет конкурентное добавление
func BenchmarkSlidingWindow_AddParallel(b *testing.B) {
	window := NewSlidingWindow(1*time.Second, 5*time.Minute)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			window.Add("iphone")
		}
	})
}

// измеряет скорость получения статистики
func BenchmarkSlidingWindow_GetStats(b *testing.B) {
	window := NewSlidingWindow(1*time.Second, 5*time.Minute)

	for i := 0; i < 10000; i++ {
		window.Add("iphone")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		window.GetStats()
	}
}

// измеряет смешанную нагрузку
func BenchmarkSlidingWindow_Mixed(b *testing.B) {
	window := NewSlidingWindow(1*time.Second, 5*time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			window.Add("iphone")
			window.GetStats()
		}
	})
}
