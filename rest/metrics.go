package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus struct {
	requestCount  *prometheus.CounterVec
	responseCount *prometheus.CounterVec
	latency       *prometheus.HistogramVec
}

func NewPrometheusService(config Config) (*Prometheus, error) {
	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "total_request",
			Help:      "Total number of HTTP requests",
		},
		[]string{"path"},
	)

	responseCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "total_response",
			Help:      "Total number of error HTTP requests",
		},
		[]string{"path", "code"},
	)

	latency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "request_latency",
			Help:      "Response latency in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2},
		},
		[]string{"path"},
	)

	s := &Prometheus{
		requestCount:  requestCount,
		responseCount: responseCount,
		latency:       latency,
	}

	err := prometheus.Register(s.requestCount)
	if err != nil && err.Error() != "duplicate metrics collector registration attempted" {
		return nil, err
	}

	err = prometheus.Register(s.responseCount)
	if err != nil && err.Error() != "duplicate metrics collector registration attempted" {
		return nil, err
	}

	err = prometheus.Register(s.latency)
	if err != nil && err.Error() != "duplicate metrics collector registration attempted" {
		return nil, err
	}

	return s, nil
}

func (s *Prometheus) RequestMetricsMiddleware(path string, next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			delta := time.Since(start).Seconds()
			s.latency.WithLabelValues(path).Observe(delta)
		}()
		s.requestCount.WithLabelValues(path).Inc()

		writer := NewLoggingResponseWriter(w)
		next.ServeHTTP(writer, r)

		status := writer.Code()
		if status >= http.StatusMultipleChoices {
			s.responseCount.WithLabelValues(path, strconv.Itoa(status)).Inc()
		}
	}
	return http.HandlerFunc(f)
}
