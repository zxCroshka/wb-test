package cache

import (
	"testing"

	"github.com/zxCroshka/wb-test/internal/domain"
)

// измеряет скорость чтения из кэша
func BenchmarkTopCache_Get(b *testing.B) {
	cache := NewTopCache()
	items := []domain.RankItem{
		{Query: "iphone", Count: 100},
		{Query: "samsung", Count: 80},
		{Query: "xiaomi", Count: 60},
	}
	cache.Set(items)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get()
	}
}

// измеряет конкурентное чтение
func BenchmarkTopCache_GetParallel(b *testing.B) {
	cache := NewTopCache()
	items := []domain.RankItem{
		{Query: "iphone", Count: 100},
		{Query: "samsung", Count: 80},
		{Query: "xiaomi", Count: 60},
	}
	cache.Set(items)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get()
		}
	})
}

// измеряет получение топа с лимитом
func BenchmarkTopCache_GetTop(b *testing.B) {
	cache := NewTopCache()
	items := make([]domain.RankItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = domain.RankItem{
			Query: "query",
			Count: int64(100 - i),
		}
	}
	cache.Set(items)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetTop(10)
	}
}

// измеряет обновление кэша из статистики
func BenchmarkTopCache_UpdateFromStats(b *testing.B) {
	cache := NewTopCache()
	stats := make(map[string]int64)
	for i := 0; i < 1000; i++ {
		stats["query"] = int64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.UpdateFromStats(stats, 100)
	}
}

// измеряет скорость установки новых данных
func BenchmarkTopCache_Set(b *testing.B) {
	cache := NewTopCache()
	items := make([]domain.RankItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = domain.RankItem{Query: "q", Count: int64(i)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(items)
	}
}
