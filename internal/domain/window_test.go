package domain

import (
	"sync"
	"testing"
	"time"
)

func TestSlidingWindow_AddAndGetStats(t *testing.T) {
	window := NewSlidingWindow(1*time.Second, 5*time.Second)

	window.Add("iphone")
	window.Add("iphone")
	window.Add("samsung")

	stats := window.GetStats()

	if stats["iphone"] != 2 {
		t.Errorf("Expected iphone=2, got %d", stats["iphone"])
	}
	if stats["samsung"] != 1 {
		t.Errorf("Expected samsung=1, got %d", stats["samsung"])
	}
}

func TestSlidingWindow_Expiration(t *testing.T) {
	window := NewSlidingWindow(1*time.Second, 3*time.Second)

	window.Add("old_query")
	time.Sleep(4 * time.Second)

	stats := window.GetStats()
	if len(stats) != 0 {
		t.Errorf("Expected empty stats after expiration, got %v", stats)
	}
}

func TestSlidingWindow_GetActiveBucketsCount(t *testing.T) {
	window := NewSlidingWindow(1*time.Second, 5*time.Second)

	window.Add("test")
	count := window.GetActiveBucketsCount()
	if count != 1 {
		t.Errorf("Expected 1 active bucket, got %d", count)
	}
}

func TestSlidingWindow_Concurrent(t *testing.T) {
	window := NewSlidingWindow(1*time.Second, 5*time.Second)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			window.Add("test")
		}()
	}

	wg.Wait()
	stats := window.GetStats()
	if stats["test"] != 100 {
		t.Errorf("Expected 100, got %d", stats["test"])
	}
}
