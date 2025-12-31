//go:build examples

/*
Example: Using zen-sdk/pkg/leader in a controller

This example shows how to use zen-sdk/pkg/leader to enable
leader election in your controller-runtime based operator.
*/

package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kube-zen/zen-sdk/pkg/leader"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	var enableLeaderElection bool
	var leaderElectionNamespace string

	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager.")
	flag.StringVar(&leaderElectionNamespace, "leader-election-namespace", "",
		"Namespace for leader election lease.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Configure leader election using zen-sdk
	leaderOpts := leader.Options{
		LeaseName:  "my-controller-leader-election",
		Enable:     enableLeaderElection,
		Namespace:  leaderElectionNamespace,
		// Use defaults for LeaseDuration, RenewDeadline, RetryPeriod
	}

	// Create manager with leader election configured
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	}, leader.Setup(leaderOpts))
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

