package detector

import (
	"sync"

	"golang.org/x/time/rate"
)

type AnomalyDetector struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rps      int
	burst    int
}

func NewAnomalyDetector(rps int) *AnomalyDetector {
	return &AnomalyDetector{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    rps,
	}
}

func NewAnomalyDetectorWithBurst(rps, burst int) *AnomalyDetector {
	return &AnomalyDetector{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

func (d *AnomalyDetector) IsAnomaly(query, userID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := userID + ":" + query

	limiter, exists := d.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(d.rps), d.burst)
		d.limiters[key] = limiter
	}

	return !limiter.Allow()
}
