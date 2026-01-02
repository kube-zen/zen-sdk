/*
Copyright 2025 Kube-ZEN Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// HTTPMetrics provides HTTP request metrics (counter, latency histogram, inflight gauge)
type HTTPMetrics struct {
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestsInFlight *prometheus.GaugeVec
	labelExtractor   LabelExtractor
}

// LabelExtractor extracts label values from HTTP request
type LabelExtractor func(*http.Request) []string

// HTTPMetricsConfig configures HTTP metrics
type HTTPMetricsConfig struct {
	// Component name (for metric labels)
	Component string

	// Metric name prefix (default: "zen_http")
	Prefix string

	// Duration buckets for latency histogram (default: exponential buckets 25ms to 3.2s)
	DurationBuckets []float64

	// Additional labels beyond route, method, code
	ExtraLabels []string

	// LabelExtractor extracts values for ExtraLabels from request (optional)
	// Must return values in same order as ExtraLabels
	LabelExtractor LabelExtractor

	// Registry to register metrics (nil = use controller-runtime metrics.Registry)
	Registry prometheus.Registerer
}

// NewHTTPMetrics creates HTTP metrics with the given configuration
func NewHTTPMetrics(config HTTPMetricsConfig) (*HTTPMetrics, error) { //nolint:gocritic // hugeParam: config is intentionally passed by value for immutability
	prefix := config.Prefix
	if prefix == "" {
		prefix = "zen_http"
	}

	buckets := config.DurationBuckets
	if buckets == nil {
		// Default buckets: 25ms to 3.2s (similar to zen-back)
		buckets = []float64{25, 50, 100, 200, 400, 800, 1600, 3200}
	}

	// Standard labels: route, method, code (component is ConstLabel, not a variable label)
	labels := []string{"route", "method", "code"}
	labels = append(labels, config.ExtraLabels...)

	registry := config.Registry
	if registry == nil {
		registry = metrics.Registry
	}

	// Component as ConstLabel (reduces cardinality)
	constLabels := prometheus.Labels{}
	if config.Component != "" {
		constLabels["component"] = config.Component
	}

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        prefix + "_requests_total",
			Help:        "Total number of HTTP requests",
			ConstLabels: constLabels,
		},
		labels,
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        prefix + "_request_duration_ms",
			Help:        "HTTP request duration in milliseconds",
			Buckets:     buckets,
			ConstLabels: constLabels,
		},
		labels,
	)

	// Inflight requests use route label only
	inflightLabels := []string{"route"}
	requestsInFlight := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        prefix + "_requests_inflight",
			Help:        "Current number of HTTP requests being processed",
			ConstLabels: constLabels,
		},
		inflightLabels,
	)

	// Register metrics
	if err := registry.Register(requestsTotal); err != nil {
		return nil, err
	}
	if err := registry.Register(requestDuration); err != nil {
		return nil, err
	}
	if err := registry.Register(requestsInFlight); err != nil {
		return nil, err
	}

	return &HTTPMetrics{
		requestsTotal:    requestsTotal,
		requestDuration:  requestDuration,
		requestsInFlight: requestsInFlight,
		labelExtractor:   config.LabelExtractor,
	}, nil
}

// Middleware returns an HTTP middleware that records metrics
func (hm *HTTPMetrics) Middleware(route string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Increment inflight
			hm.requestsInFlight.WithLabelValues(route).Inc()
			defer hm.requestsInFlight.WithLabelValues(route).Dec()

			// Wrap ResponseWriter to capture status code
			ww := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			start := time.Now()

			// Execute handler
			next.ServeHTTP(ww, r)

			// Record metrics
			duration := time.Since(start)
			code := strconv.Itoa(ww.statusCode)
			method := r.Method

			// Build label values: route, method, code, then extra labels
			labelValues := []string{route, method, code}
			if hm.labelExtractor != nil {
				extraValues := hm.labelExtractor(r)
				labelValues = append(labelValues, extraValues...)
			}

			// Record metrics
			hm.requestsTotal.WithLabelValues(labelValues...).Inc()
			hm.requestDuration.WithLabelValues(labelValues...).Observe(float64(duration.Milliseconds()))
		})
	}
}

// statusResponseWriter wraps http.ResponseWriter to capture status code
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
