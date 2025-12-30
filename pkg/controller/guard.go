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

// Package controller provides controller utilities.
// DEPRECATED: LeaderGuard is deprecated. Use controller-runtime's built-in leader election
// via zen-sdk/pkg/leader.ApplyRequiredLeaderElection() instead.
// This package is kept for backward compatibility only and will be removed in a future version.
package controller

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// AnnotationRole is deprecated. Do not use.
	// DEPRECATED: zen-lead no longer mutates pods with leader annotations.
	// Use controller-runtime's built-in leader election instead.
	AnnotationRole = "zen-lead/role"
	// RoleLeader is deprecated. Do not use.
	RoleLeader = "leader"
)

// LeaderGuard wraps a reconcile.Reconciler to prevent "Split Brain" scenarios.
// DEPRECATED: This approach relies on pod annotations set by zen-lead, which is incompatible
// with zen-lead's Day-0 "no pod mutation" contract. Use controller-runtime's built-in leader
// election via zen-sdk/pkg/leader.ApplyRequiredLeaderElection() instead.
//
// This guard ensures only the leader pod processes reconciliation events,
// while follower pods wait and requeue. This prevents duplicate work and
// ensures consistency across replicas.
type LeaderGuard struct {
	client        client.Client
	log           logr.Logger
	isLeaderCache bool
	mu            sync.RWMutex
	podName       string
	podNamespace  string
}

// NewLeaderGuard creates a new LeaderGuard instance.
// DEPRECATED: Use controller-runtime's built-in leader election instead.
// It reads POD_NAME and POD_NAMESPACE from environment variables.
// If POD_NAME is empty (e.g., running outside a pod), it defaults to
// assuming leader status (returns true for IsLeader checks).
func NewLeaderGuard(client client.Client, logger logr.Logger) *LeaderGuard {
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = os.Getenv("HOSTNAME")
	}

	podNamespace := os.Getenv("POD_NAMESPACE")
	if podNamespace == "" {
		// Try to read from service account namespace file
		if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
			podNamespace = string(data)
		}
	}

	return &LeaderGuard{
		client:       client,
		log:          logger,
		podName:      podName,
		podNamespace: podNamespace,
	}
}

// Wrap wraps a reconcile.Reconciler with leader guard logic.
// DEPRECATED: Use controller-runtime's built-in leader election instead.
// Only the leader pod processes reconciliation events; all
// followers requeue and wait.
func (lg *LeaderGuard) Wrap(inner reconcile.Reconciler) reconcile.Reconciler {
	return reconcile.Func(func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		// Fast Path: Check cache with read lock
		lg.mu.RLock()
		isLeader := lg.isLeaderCache
		lg.mu.RUnlock()

		if isLeader {
			// We are the leader, execute inner reconciler
			return inner.Reconcile(ctx, req)
		}

		// Slow Path: Check pod annotations (requires write lock for cache update)
		lg.mu.Lock()
		defer lg.mu.Unlock()

		// Double-check after acquiring write lock (another goroutine might have updated)
		if lg.isLeaderCache {
			return inner.Reconcile(ctx, req)
		}

		// If running outside a pod (local dev), assume leader
		if lg.podName == "" {
			lg.isLeaderCache = true
			return inner.Reconcile(ctx, req)
		}

		// Fetch current pod to check leader status
		pod := &corev1.Pod{}
		key := types.NamespacedName{
			Name:      lg.podName,
			Namespace: lg.podNamespace,
		}

		if err := lg.client.Get(ctx, key, pod); err != nil {
			// If we can't fetch the pod, log error but don't fail
			// This might happen during pod startup or API server issues
			lg.log.Error(err, "Failed to fetch pod for leader check, assuming follower",
				"pod", lg.podName,
				"namespace", lg.podNamespace,
			)
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}

		// Check zen-lead/role annotation (DEPRECATED: zen-lead no longer sets this)
		role, exists := pod.Annotations[AnnotationRole]
		if !exists {
			// No annotation means not participating in leader election (or zen-lead not running)
			lg.log.V(4).Info("No zen-lead/role annotation found, assuming follower (DEPRECATED: use controller-runtime leader election)",
				"pod", lg.podName,
			)
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}

		if role == RoleLeader {
			// We are the leader! Update cache and execute inner reconciler
			lg.isLeaderCache = true
			lg.log.Info("Elected as leader, processing reconciliation (DEPRECATED: use controller-runtime leader election)",
				"pod", lg.podName,
			)
			return inner.Reconcile(ctx, req)
		}

		// We are a follower, log and requeue
		lg.log.V(4).Info("Not leader, requeuing (DEPRECATED: use controller-runtime leader election)",
			"pod", lg.podName,
			"role", role,
		)
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	})
}
