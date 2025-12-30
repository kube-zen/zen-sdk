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
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Recorder records metrics for Kubernetes controllers
type Recorder struct {
	componentName string
	
	// Reconciliation metrics
	reconciliationsTotal *prometheus.CounterVec
	reconciliationsDuration *prometheus.HistogramVec
	
	// Error metrics
	errorsTotal *prometheus.CounterVec
	
	// Custom metrics can be added here
}

var (
	collectorsRegistered bool
	collectorsMu         sync.Mutex
)

// NewRecorder creates a new metrics recorder for a component
func NewRecorder(componentName string) *Recorder {
	recorder := &Recorder{
		componentName: componentName,
	}
	
	// Register standard Prometheus collectors (only once)
	collectorsMu.Lock()
	if !collectorsRegistered {
		metrics.Registry.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
		collectorsRegistered = true
	}
	collectorsMu.Unlock()
	
	// Reconciliation counter
	recorder.reconciliationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "zen_reconciliations_total",
			Help: "Total number of reconciliations",
			ConstLabels: prometheus.Labels{
				"component": componentName,
			},
		},
		[]string{"result"}, // "success", "error"
	)
	
	// Reconciliation duration histogram
	recorder.reconciliationsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "zen_reconciliation_duration_seconds",
			Help: "Duration of reconciliations in seconds",
			ConstLabels: prometheus.Labels{
				"component": componentName,
			},
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		},
		[]string{"result"},
	)
	
	// Error counter
	recorder.errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "zen_errors_total",
			Help: "Total number of errors",
			ConstLabels: prometheus.Labels{
				"component": componentName,
			},
		},
		[]string{"type"}, // "reconciliation", "webhook", etc.
	)
	
	// Register metrics
	metrics.Registry.MustRegister(
		recorder.reconciliationsTotal,
		recorder.reconciliationsDuration,
		recorder.errorsTotal,
	)
	
	return recorder
}

// RecordReconciliation records a reconciliation attempt
func (r *Recorder) RecordReconciliation(result string, durationSeconds float64) {
	r.reconciliationsTotal.WithLabelValues(result).Inc()
	r.reconciliationsDuration.WithLabelValues(result).Observe(durationSeconds)
}

// RecordError records an error
func (r *Recorder) RecordError(errorType string) {
	r.errorsTotal.WithLabelValues(errorType).Inc()
}

// RecordReconciliationSuccess is a convenience method for successful reconciliations
func (r *Recorder) RecordReconciliationSuccess(durationSeconds float64) {
	r.RecordReconciliation("success", durationSeconds)
}

// RecordReconciliationError is a convenience method for failed reconciliations
func (r *Recorder) RecordReconciliationError(durationSeconds float64) {
	r.RecordReconciliation("error", durationSeconds)
	r.RecordError("reconciliation")
}

