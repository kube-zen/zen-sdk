/*
Example: Using zen-sdk/pkg/metrics in a controller

This example shows how to use zen-sdk/pkg/metrics to record
metrics in your controller.
*/

package main

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kube-zen/zen-sdk/pkg/metrics"
)

// ExampleReconciler shows how to use metrics in a reconciler
type ExampleReconciler struct {
	recorder *metrics.Recorder
}

func (r *ExampleReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	start := time.Now()
	
	// Your reconciliation logic here
	// ...
	
	// Record metrics
	duration := time.Since(start).Seconds()
	r.recorder.RecordReconciliationSuccess(duration)
	
	return reconcile.Result{}, nil
}

func ExampleUsage() {
	// Create metrics recorder
	recorder := metrics.NewRecorder("my-controller")
	
	// Use in reconciler
	reconciler := &ExampleReconciler{
		recorder: recorder,
	}
	
	// Record errors
	recorder.RecordError("reconciliation")
	
	// Record reconciliation with custom result
	recorder.RecordReconciliation("success", 0.5)
	recorder.RecordReconciliation("error", 1.0)
}

