package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// inspired by https://github.com/zsais/go-gin-prometheus

var (
	defaultMetricPath = "/metrics"
	subsystem         = "gin"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method", "handler", "host", "url"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds.",
		},
		[]string{"code", "method", "url"},
	)
	requestSize = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
		},
	)
	responseSize = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
		},
	)
	calculationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "calc", // Add observability for the calculations done by the Worker
			Name:      "calculation_duration_seconds",
			Help:      "Duration of calculations in seconds.",
		},
		[]string{"docid"},
	)
)

// init runs before main() is evaluated - register our metrics with prometheus
func init() {
	prometheus.MustRegister(requestsTotal, requestDuration, requestSize, responseSize, calculationDuration)
}

// prometheusMiddleware is a gin middleware function that instruments each request made
// to the API server
func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Don't instrument the metrics path
		if c.Request.URL.Path == defaultMetricPath {
			c.Next()
			return
		}

		start := time.Now()
		reqSize := computeApproximateRequestSize(c.Request)

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()
		resSize := float64(c.Writer.Size())
		url := c.FullPath() // Record the template name that was called - i.e. /jobs/:id

		requestsTotal.WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host, url).Inc()
		requestDuration.WithLabelValues(status, c.Request.Method, url).Observe(duration)
		requestSize.Observe(float64(reqSize))
		responseSize.Observe(resSize)
	}
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
