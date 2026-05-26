package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	searchEventsTotal   prometheus.Counter
	blockedQueriesTotal prometheus.Counter
	anomalyQueriesTotal prometheus.Counter
	cacheHitsTotal      prometheus.Counter
	cacheMissesTotal    prometheus.Counter

	topCacheSize        prometheus.Gauge
	windowTotalQueries  prometheus.Gauge
	windowUniqueQueries prometheus.Gauge
	windowActiveBuckets prometheus.Gauge

	aggregationDuration prometheus.Histogram

	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestSize     *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	m := &Metrics{
		searchEventsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "search_events_total",
			Help: "Total number of search events processed",
		}),
		blockedQueriesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "blocked_queries_total",
			Help: "Total number of queries blocked by stoplist",
		}),
		anomalyQueriesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "anomaly_queries_total",
			Help: "Total number of queries detected as anomalies",
		}),
		cacheHitsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		}),
		cacheMissesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		}),

		topCacheSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "top_cache_size",
			Help: "Current size of top cache",
		}),
		windowTotalQueries: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "window_total_queries",
			Help: "Total number of queries in sliding window",
		}),
		windowUniqueQueries: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "window_unique_queries",
			Help: "Number of unique queries in sliding window",
		}),
		windowActiveBuckets: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "window_active_buckets",
			Help: "Number of active buckets in sliding window",
		}),

		aggregationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "aggregation_duration_seconds",
			Help:    "Time spent aggregating top queries",
			Buckets: prometheus.DefBuckets,
		}),

		httpRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		}, []string{"method", "endpoint", "status"}),

		httpRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"method", "endpoint"}),

		httpRequestSize: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000},
		}, []string{"method", "endpoint"}),

		httpResponseSize: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000},
		}, []string{"method", "endpoint"}),
	}

	return m
}

func (m *Metrics) IncSearchEvents()         { m.searchEventsTotal.Inc() }
func (m *Metrics) IncBlockedQueries()       { m.blockedQueriesTotal.Inc() }
func (m *Metrics) IncAnomalyQueries()       { m.anomalyQueriesTotal.Inc() }
func (m *Metrics) IncCacheHits()            { m.cacheHitsTotal.Inc() }
func (m *Metrics) IncCacheMisses()          { m.cacheMissesTotal.Inc() }
func (m *Metrics) SetTopCacheSize(size int) { m.topCacheSize.Set(float64(size)) }

func (m *Metrics) SetWindowStats(totalQueries, uniqueQueries int64, activeBuckets int) {
	m.windowTotalQueries.Set(float64(totalQueries))
	m.windowUniqueQueries.Set(float64(uniqueQueries))
	m.windowActiveBuckets.Set(float64(activeBuckets))
}

func (m *Metrics) ObserveAggregationDuration(duration float64) {
	m.aggregationDuration.Observe(duration)
}

func (m *Metrics) RecordHTTPRequest(method, endpoint string, status int) {
	m.httpRequestsTotal.WithLabelValues(method, endpoint, strconv.Itoa(status)).Inc()
}

func (m *Metrics) RecordHTTPDuration(method, endpoint string, duration float64) {
	m.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

func (m *Metrics) RecordHTTPRequestSize(method, endpoint string, size float64) {
	m.httpRequestSize.WithLabelValues(method, endpoint).Observe(size)
}

func (m *Metrics) RecordHTTPResponseSize(method, endpoint string, size float64) {
	m.httpResponseSize.WithLabelValues(method, endpoint).Observe(size)
}
