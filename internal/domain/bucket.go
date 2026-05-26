package domain

import "sync"

type Bucket struct {
	mu     sync.RWMutex
	counts map[string]int64
}

func NewBucket() *Bucket {
	return &Bucket{
		counts: make(map[string]int64),
	}
}

func (b *Bucket) Increment(query string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.counts[query]++
}

func (b *Bucket) GetCounts() map[string]int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	res := make(map[string]int64)
	for k, v := range b.counts {
		res[k] = v
	}
	return res
}

func (b *Bucket) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.counts = make(map[string]int64)
}
