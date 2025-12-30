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

package leader

import (
	"context"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// AnnotationRole is the annotation key set by zen-lead controller
	AnnotationRole = "zen-lead/role"
	// RoleLeader indicates this pod is the leader
	RoleLeader = "leader"
	// RoleFollower indicates this pod is a follower
	RoleFollower = "follower"
	// DefaultCheckInterval is the default interval for checking leader status
	DefaultCheckInterval = 5 * time.Second
)

// Watcher watches zen-lead annotations to determine leader status
// This is used when ha.mode=external (zen-lead controller manages leader election)
type Watcher struct {
	client       client.Client
	podName      string
	podNamespace string
	checkInterval time.Duration
	isLeader     bool
	onLeaderChange func(bool)
}

// NewWatcher creates a new external leader watcher for zen-lead
func NewWatcher(client client.Client, onLeaderChange func(bool)) (*Watcher, error) {
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = os.Getenv("HOSTNAME")
	}
	if podName == "" {
		return nil, fmt.Errorf("POD_NAME or HOSTNAME environment variable not set")
	}

	podNamespace := os.Getenv("POD_NAMESPACE")
	if podNamespace == "" {
		// Try to read from service account namespace
		if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
			podNamespace = string(data)
		}
	}
	if podNamespace == "" {
		return nil, fmt.Errorf("POD_NAMESPACE environment variable not set and cannot read from service account")
	}

	checkInterval := DefaultCheckInterval
	if intervalStr := os.Getenv("ZEN_LEAD_CHECK_INTERVAL"); intervalStr != "" {
		if parsed, err := time.ParseDuration(intervalStr); err == nil && parsed > 0 {
			checkInterval = parsed
		}
	}

	return &Watcher{
		client:        client,
		podName:       podName,
		podNamespace:  podNamespace,
		checkInterval: checkInterval,
		onLeaderChange: onLeaderChange,
	}, nil
}

// IsLeader checks if the current pod is the leader by reading zen-lead/role annotation
func (w *Watcher) IsLeader(ctx context.Context) (bool, error) {
	pod := &corev1.Pod{}
	key := types.NamespacedName{
		Name:      w.podName,
		Namespace: w.podNamespace,
	}

	if err := w.client.Get(ctx, key, pod); err != nil {
		return false, fmt.Errorf("failed to get pod %s/%s: %w", w.podNamespace, w.podName, err)
	}

	// Check zen-lead/role annotation
	role, exists := pod.Annotations[AnnotationRole]
	if !exists {
		// No annotation means not participating in leader election (or zen-lead not running)
		return false, nil
	}

	return role == RoleLeader, nil
}

// Watch watches for leader status changes and calls the callback
func (w *Watcher) Watch(ctx context.Context) error {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	var lastLeaderState bool
	firstCheck := true

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			isLeader, err := w.IsLeader(ctx)
			if err != nil {
				// Log error but continue watching
				continue
			}

			// Call callback on first check or when state changes
			if firstCheck || isLeader != lastLeaderState {
				w.isLeader = isLeader
				if w.onLeaderChange != nil {
					w.onLeaderChange(isLeader)
				}
				lastLeaderState = isLeader
				firstCheck = false
			}
		}
	}
}

// GetIsLeader returns the cached leader status
func (w *Watcher) GetIsLeader() bool {
	return w.isLeader
}

