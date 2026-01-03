//go:build examples

/*
Example: Hello Controller - Minimal Controller Example

This is a minimal, copy-paste ready controller example that demonstrates
basic controller-runtime usage with zen-sdk packages. It watches ConfigMaps
and logs when they are created, updated, or deleted.

To use this example:
1. Copy this file to your project
2. Run: go mod init hello-controller
3. Run: go get github.com/kube-zen/zen-sdk@latest
4. Run: go get sigs.k8s.io/controller-runtime@latest
5. Run: go run hello_controller.go

The controller will:
- Watch ConfigMaps in all namespaces
- Log when ConfigMaps are created/updated/deleted
- Use zen-sdk for logging, metrics, lifecycle, and health checks
- Compile in under 2 minutes
*/

package main

import (
	"context"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kube-zen/zen-sdk/pkg/health"
	"github.com/kube-zen/zen-sdk/pkg/lifecycle"
	"github.com/kube-zen/zen-sdk/pkg/logging"
	"github.com/kube-zen/zen-sdk/pkg/metrics"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	// Initialize zen-sdk logger
	logger := logging.NewLogger("hello-controller")
	logger.Info("Starting Hello Controller")

	// Get Kubernetes config (uses KUBECONFIG env var or ~/.kube/config)
	cfg := ctrl.GetConfigOrDie()

	// Create controller manager
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		// Metrics and health endpoints
		MetricsBindAddress: ":8080",
		HealthProbeBindAddress: ":8081",
	})
	if err != nil {
		logger.Error(err, "Failed to create manager")
		os.Exit(1)
	}

	// Setup metrics recorder
	metricsRecorder := metrics.NewRecorder("hello-controller")

	// Setup health checks
	informerSyncChecker := health.NewInformerSyncChecker(
		func(ctx context.Context) (map[string]bool, error) {
			return mgr.GetCache().WaitForCacheSync(ctx), nil
		},
	)
	if err := mgr.AddHealthzCheck("healthz", informerSyncChecker.LivenessCheck); err != nil {
		logger.Error(err, "Failed to add healthz check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", informerSyncChecker.ReadinessCheck); err != nil {
		logger.Error(err, "Failed to add readyz check")
		os.Exit(1)
	}

	// Create reconciler
	reconciler := &ConfigMapReconciler{
		Client:   mgr.GetClient(),
		Logger:   logger,
		Metrics:  metricsRecorder,
	}

	// Setup controller
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(reconciler); err != nil {
		logger.Error(err, "Failed to create controller")
		os.Exit(1)
	}

	// Setup graceful shutdown
	ctx, cancel := lifecycle.ShutdownContext(context.Background(), "hello-controller")
	defer cancel()

	// Start manager in goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Info("Starting manager")
		if err := mgr.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("Shutdown signal received, stopping controller")

	// Give manager time to shut down gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Wait for manager to stop or timeout
	select {
	case err := <-errChan:
		if err != nil {
			logger.Error(err, "Manager stopped with error")
		} else {
			logger.Info("Manager stopped gracefully")
		}
	case <-shutdownCtx.Done():
		logger.Info("Shutdown timeout reached")
	}
}

// ConfigMapReconciler reconciles ConfigMap objects
type ConfigMapReconciler struct {
	client.Client
	Logger  *logging.Logger
	Metrics *metrics.Recorder
}

// Reconcile is called whenever a ConfigMap is created, updated, or deleted
func (r *ConfigMapReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	start := time.Now()

	// Get the ConfigMap
	configMap := &corev1.ConfigMap{}
	if err := r.Get(ctx, req.NamespacedName, configMap); err != nil {
		// ConfigMap was deleted
		if client.IgnoreNotFound(err) == nil {
			r.Logger.Info("ConfigMap deleted",
				logging.String("namespace", req.Namespace),
				logging.String("name", req.Name),
			)
			r.Metrics.RecordReconciliationSuccess(time.Since(start).Seconds())
			return reconcile.Result{}, nil
		}
		// Error getting ConfigMap
		r.Logger.Error(err, "Failed to get ConfigMap",
			logging.String("namespace", req.Namespace),
			logging.String("name", req.Name),
		)
		r.Metrics.RecordReconciliationError(time.Since(start).Seconds())
		return reconcile.Result{}, err
	}

	// ConfigMap exists - log it
	r.Logger.Info("ConfigMap reconciled",
		logging.String("namespace", configMap.Namespace),
		logging.String("name", configMap.Name),
		logging.Int("data_keys", len(configMap.Data)),
		logging.String("resource_version", configMap.ResourceVersion),
	)

	// Record success metric
	r.Metrics.RecordReconciliationSuccess(time.Since(start).Seconds())

	return reconcile.Result{}, nil
}

