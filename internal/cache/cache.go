package cache

import (
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/zxCroshka/wb-test/internal/domain"
)

type TopCache struct {
	ptr       atomic.Pointer[[]domain.RankItem]
	updatedAt atomic.Int64
}

func NewTopCache() *TopCache {
	tc := &TopCache{}
	empty := make([]domain.RankItem, 0)
	tc.ptr.Store(&empty)
	tc.updatedAt.Store(time.Now().Unix())
	return tc
}

func (c *TopCache) Get() []domain.RankItem {
	ptr := c.ptr.Load()
	if ptr == nil {
		return []domain.RankItem{}
	}
	return *ptr
}

func (c *TopCache) GetTop(limit int) ([]domain.RankItem, error) {
	allTop := c.Get()
	if limit <= 0 {
		return []domain.RankItem{}, fmt.Errorf("limit cant be less than 0")
	}
	if limit >= len(allTop) {
		result := make([]domain.RankItem, len(allTop))
		copy(result, allTop)
		return result, nil
	}

	result := make([]domain.RankItem, limit)
	copy(result, allTop[:limit])
	return result, nil
}

func (c *TopCache) Set(items []domain.RankItem) {
	if items == nil {
		items = make([]domain.RankItem, 0)
	}
	c.ptr.Store(&items)
	c.updatedAt.Store(time.Now().Unix())
}

func (c *TopCache) UpdateFromStats(stats map[string]int64, maxSize int) {
	items := make([]domain.RankItem, 0, len(stats))
	for query, count := range stats {
		items = append(items, domain.RankItem{
			Query: query,
			Count: count,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Query < items[j].Query
		}
		return items[i].Count > items[j].Count
	})

	if len(items) > maxSize {
		items = items[:maxSize]
	}

	c.Set(items)
}

func (c *TopCache) Size() int {
	ptr := c.ptr.Load()
	if ptr == nil {
		return 0
	}
	return len(*ptr)
}
