package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zxCroshka/wb-test/internal/metrics"
)

func metricsMiddleware(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// НЕ перехватываем ResponseWriter, чтобы не мешать Gin
		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		m.RecordHTTPRequest(method, path, status)
		m.RecordHTTPDuration(method, path, duration)

		// Размеры можно брать из c.Writer.Size()
		m.RecordHTTPRequestSize(method, path, float64(c.Request.ContentLength))
		m.RecordHTTPResponseSize(method, path, float64(c.Writer.Size()))
	}
}
