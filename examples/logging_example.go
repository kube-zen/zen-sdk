//go:build examples

/*
Example: Using zen-sdk/pkg/logging in a controller

This example shows how to use zen-sdk/pkg/logging for
structured logging in your controller.
*/

package main

import (
	"github.com/kube-zen/zen-sdk/pkg/logging"
)

func ExampleUsage() {
	// Create logger with component name
	logger := logging.NewLogger("my-controller")
	
	// Basic logging
	logger.Info("Controller started")
	logger.Error(nil, "Controller error")
	
	// With context
	logger.WithField("namespace", "default").Info("Processing namespace")
	logger.WithField("resource", "my-resource").Info("Processing resource")
	
	// With multiple fields
	fields := map[string]interface{}{
		"namespace": "default",
		"name":      "my-resource",
		"action":    "reconcile",
	}
	logger.WithFields(fields).Info("Reconciling resource")
	
	// With component context
	componentLogger := logger.WithComponent("reconciler")
	componentLogger.Info("Reconciler started")
}

