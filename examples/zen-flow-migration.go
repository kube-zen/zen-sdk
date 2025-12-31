//go:build examples

/*
Example: Migrating zen-flow to use zen-sdk

This example shows how zen-flow would migrate from custom
leader election code to using zen-sdk/pkg/leader.
*/

package main

import (
	"flag"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/kube-zen/zen-flow/pkg/api/v1alpha1"
	"github.com/kube-zen/zen-flow/pkg/controller"

	// Import zen-sdk packages
	"github.com/kube-zen/zen-sdk/pkg/leader"
	"github.com/kube-zen/zen-sdk/pkg/logging"
	"github.com/kube-zen/zen-sdk/pkg/metrics"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var namespace string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager.")
	flag.StringVar(&namespace, "namespace", "", "Namespace to watch (empty = all namespaces)")
	
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// Use zen-sdk logging
	logger := logging.NewLogger("zen-flow")
	logger.Info("Starting zen-flow controller")

	// Setup metrics using zen-sdk
	metricsRecorder := metrics.NewRecorder("zen-flow")
	logger.Info("Metrics recorder initialized")

	// Configure leader election using zen-sdk
	leaderOpts := leader.Options{
		LeaseName:  "zen-flow-controller-leader-election",
		Enable:     enableLeaderElection,
		Namespace:  namespace,
		// Uses defaults: 15s lease, 10s renew, 2s retry
	}

	// Create manager with leader election configured via zen-sdk
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
	}, leader.Setup(leaderOpts))
	if err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Setup controller with metrics recorder
	if err := controller.SetupController(mgr, 1, metricsRecorder, nil); err != nil {
		logger.Error(err, "unable to setup controller")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	logger.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "problem running manager")
		os.Exit(1)
	}
}

/*
BEFORE (zen-flow/pkg/controller/manager.go):

func ManagerOptions(namespace string, enableLeaderElection bool) ctrl.Options {
    opts := ctrl.Options{
        Scheme:                  nil,
        LeaderElection:          enableLeaderElection,
        LeaderElectionID:        "zen-flow-controller-leader-election",
        LeaderElectionNamespace: namespace,
        LeaseDuration:           func() *time.Duration { d := 15 * time.Second; return &d }(),
        RenewDeadline:           func() *time.Duration { d := 10 * time.Second; return &d }(),
        RetryPeriod:             func() *time.Duration { d := 2 * time.Second; return &d }(),
    }
    return opts
}

AFTER (Using zen-sdk):

import "github.com/kube-zen/zen-sdk/pkg/leader"

leaderOpts := leader.Options{
    LeaseName:  "zen-flow-controller-leader-election",
    Enable:     enableLeaderElection,
    Namespace:  namespace,
}
mgr, err := ctrl.NewManager(cfg, ctrl.Options{}, leader.Setup(leaderOpts))

Benefits:
- 10+ lines of code removed
- Consistent with other tools
- Easier to maintain
- Well-tested SDK code
*/

