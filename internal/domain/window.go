package domain

import (
	"sync"
	"time"
)

type SlidingWindow struct {
	buckets    []*Bucket
	curIdx     int
	lastUpdate time.Time
	bucketSize time.Duration
	windowSize time.Duration

	mu sync.RWMutex
}

func NewSlidingWindow(bucketSize, windowSize time.Duration) *SlidingWindow {
	bucketCount := int(windowSize / bucketSize)
	buckets := make([]*Bucket, bucketCount)
	for i := 0; i < bucketCount; i++ {
		buckets[i] = NewBucket()
	}
	return &SlidingWindow{
		buckets:    buckets,
		curIdx:     0,
		lastUpdate: time.Now(),
		bucketSize: bucketSize,
		windowSize: windowSize,
	}
}

func (sw *SlidingWindow) Add(query string) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.advance(time.Now())
	sw.buckets[sw.curIdx].Increment(query)
}

func (sw *SlidingWindow) advance(now time.Time) {
	elapsed := now.Sub(sw.lastUpdate)

	bucketsToAdvance := int(elapsed / sw.bucketSize)
	if bucketsToAdvance <= 0 {
		return
	}
	if bucketsToAdvance > len(sw.buckets) {
		bucketsToAdvance = len(sw.buckets)
	}

	for i := 0; i < bucketsToAdvance; i++ {
		sw.curIdx = (sw.curIdx + 1) % len(sw.buckets)
		sw.buckets[sw.curIdx].Reset()
	}
	sw.lastUpdate = now
}

func (sw *SlidingWindow) GetStats() map[string]int64 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	sw.advance(time.Now())

	result := make(map[string]int64)
	for _, bucket := range sw.buckets {
		bucket.mu.RLock()
		for q, c := range bucket.counts {
			result[q] += c
		}
		bucket.mu.RUnlock()
	}
	return result
}
func (sw *SlidingWindow) GetActiveBucketsCount() int {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	count := 0
	for _, bucket := range sw.buckets {
		bucket.mu.RLock()
		if len(bucket.counts) > 0 {
			count++
		}
		bucket.mu.RUnlock()
	}
	return count
}
