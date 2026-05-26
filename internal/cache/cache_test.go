package cache

import (
	"sync"
	"testing"

	"github.com/zxCroshka/wb-test/internal/domain"
)

func TestTopCache_SetAndGet(t *testing.T) {
	cache := NewTopCache()
	items := []domain.RankItem{{Query: "test", Count: 1}}

	cache.Set(items)
	result := cache.Get()

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}
}

func TestTopCache_GetTop(t *testing.T) {
	cache := NewTopCache()
	items := []domain.RankItem{
		{Query: "a", Count: 10},
		{Query: "b", Count: 9},
		{Query: "c", Count: 8},
	}
	cache.Set(items)

	top, err := cache.GetTop(2)
	if err != nil {
		t.Error(err)
	}
	if len(top) != 2 {
		t.Errorf("Expected 2 items, got %d", len(top))
	}
}

func TestTopCache_UpdateFromStats(t *testing.T) {
	cache := NewTopCache()
	stats := map[string]int64{
		"iphone":  100,
		"samsung": 80,
		"xiaomi":  120,
	}

	cache.UpdateFromStats(stats, 2)
	result := cache.Get()

	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}
	if result[0].Query != "xiaomi" {
		t.Errorf("Expected xiaomi first, got %s", result[0].Query)
	}
}

func TestTopCache_Concurrency(t *testing.T) {
	cache := NewTopCache()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Get()
		}()
	}
	wg.Wait()
}
