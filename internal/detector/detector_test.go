package detector

import (
	"sync"
	"testing"
)

func TestAnomalyDetector_IsAnomaly(t *testing.T) {
	d := NewAnomalyDetector(5)

	for i := 0; i < 5; i++ {
		if d.IsAnomaly("iphone", "user1") {
			t.Errorf("Request %d should not be anomaly", i+1)
		}
	}

	if !d.IsAnomaly("iphone", "user1") {
		t.Error("6th request should be anomaly")
	}
}

func TestAnomalyDetector_DifferentUsers(t *testing.T) {
	d := NewAnomalyDetector(3)

	for i := 0; i < 3; i++ {
		d.IsAnomaly("iphone", "user1")
	}

	if d.IsAnomaly("iphone", "user2") {
		t.Error("Different user should not be anomaly")
	}
}

func TestAnomalyDetector_DifferentQueries(t *testing.T) {
	d := NewAnomalyDetector(3)

	for i := 0; i < 3; i++ {
		d.IsAnomaly("iphone", "user1")
	}

	if d.IsAnomaly("samsung", "user1") {
		t.Error("Different query should not be anomaly")
	}
}

func TestAnomalyDetector_Concurrent(t *testing.T) {
	d := NewAnomalyDetector(100)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.IsAnomaly("test", "user")
		}()
	}

	wg.Wait()
}
