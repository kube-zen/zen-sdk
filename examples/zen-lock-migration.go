//go:build examples

/*
Example: Migrating zen-lock to use zen-sdk

This example shows how zen-lock would migrate from custom
code to using zen-sdk packages.
*/

package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	securityv1alpha1 "github.com/kube-zen/zen-lock/pkg/apis/security.kube-zen.io/v1alpha1"
	"github.com/kube-zen/zen-lock/pkg/controller"
	webhookpkg "github.com/kube-zen/zen-lock/pkg/webhook"

	// Import zen-sdk packages
	"github.com/kube-zen/zen-sdk/pkg/leader"
	"github.com/kube-zen/zen-sdk/pkg/logging"
	"github.com/kube-zen/zen-sdk/pkg/metrics"
	sdkwebhook "github.com/kube-zen/zen-sdk/pkg/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(securityv1alpha1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var certDir string
	var enableController bool
	var enableWebhook bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager.")
	flag.StringVar(&certDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs",
		"The directory where cert-manager injects the TLS certificates.")
	flag.BoolVar(&enableController, "enable-controller", true,
		"Enable the controller (ZenLock and Secret reconcilers).")
	flag.BoolVar(&enableWebhook, "enable-webhook", true,
		"Enable the mutating admission webhook.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// Use zen-sdk logging
	logger := logging.NewLogger("zen-lock")
	logger.Info("Starting zen-lock webhook")

	// Setup metrics using zen-sdk
	metricsRecorder := metrics.NewRecorder("zen-lock")
	logger.Info("Metrics recorder initialized")

	// Check for private key
	if os.Getenv("ZEN_LOCK_PRIVATE_KEY") == "" {
		logger.Error(nil, "ZEN_LOCK_PRIVATE_KEY not set", "error", "Private key environment variable is required")
		os.Exit(1)
	}

	// Configure leader election using zen-sdk
	leaderOpts := leader.Options{
		LeaseName: "zen-lock-webhook-leader-election",
		Enable:    enableLeaderElection,
	}

	// Create manager with leader election configured via zen-sdk
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port:    9443,
			CertDir: certDir,
		}),
		HealthProbeBindAddress: probeAddr,
	}, leader.Setup(leaderOpts))
	if err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Setup ZenLock controller (if enabled)
	if enableController {
		zenlockReconciler, err := controller.NewZenLockReconciler(mgr.GetClient(), mgr.GetScheme())
		if err != nil {
			logger.Error(err, "unable to create ZenLock reconciler")
			os.Exit(1)
		}
		if err := zenlockReconciler.SetupWithManager(mgr); err != nil {
			logger.Error(err, "unable to setup ZenLock controller")
			os.Exit(1)
		}

		secretReconciler := controller.NewSecretReconciler(mgr.GetClient(), mgr.GetScheme())
		if err := secretReconciler.SetupWithManager(mgr); err != nil {
			logger.Error(err, "unable to setup Secret controller")
			os.Exit(1)
		}
	}

	// Setup webhook (if enabled)
	if enableWebhook {
		// Use zen-sdk webhook helpers for patch generation
		// Example: webhookpkg could use sdkwebhook.GeneratePatch()
		if err := webhookpkg.SetupWebhook(mgr); err != nil {
			logger.Error(err, "unable to setup webhook")
			os.Exit(1)
		}
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
BEFORE (zen-lock/cmd/webhook/main.go):

mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme: scheme,
    Metrics: metricsserver.Options{
        BindAddress: metricsAddr,
    },
    WebhookServer: webhook.NewServer(webhook.Options{
        Port:    9443,
        CertDir: certDir,
    }),
    HealthProbeBindAddress: probeAddr,
    LeaderElection:         enableLeaderElection,
    LeaderElectionID:       "zen-lock-webhook-leader-election",
})

AFTER (Using zen-sdk):

import "github.com/kube-zen/zen-sdk/pkg/leader"

leaderOpts := leader.Options{
    LeaseName: "zen-lock-webhook-leader-election",
    Enable:    enableLeaderElection,
}

mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme: scheme,
    // ... other options ...
}, leader.Setup(leaderOpts))

Benefits:
- Consistent leader election configuration
- Can use zen-sdk/pkg/webhook for patch generation
- Can use zen-sdk/pkg/metrics for metrics
- Can use zen-sdk/pkg/logging for logging
- Single source of truth
*/

