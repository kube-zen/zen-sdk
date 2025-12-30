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
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Options configures leader election for controller-runtime Manager
type Options struct {
	// LeaseName is the name of the Lease resource used for leader election
	// Default: "<component-name>-leader-election"
	LeaseName string

	// Enable enables leader election
	// Default: false
	Enable bool

	// Namespace is the namespace where the Lease resource is created
	// If empty, uses the manager's namespace
	Namespace string

	// LeaseDuration is how long a leader holds the lease before it expires
	// Default: 15 seconds
	LeaseDuration time.Duration

	// RenewDeadline is the time to renew the lease before losing leadership
	// Default: 10 seconds (must be < LeaseDuration)
	RenewDeadline time.Duration

	// RetryPeriod is how often to retry acquiring leadership
	// Default: 2 seconds
	RetryPeriod time.Duration
}

// DefaultOptions returns default leader election options
func DefaultOptions(leaseName string) Options {
	return Options{
		LeaseName:     leaseName,
		Enable:         false,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
	}
}

// Setup configures leader election options for controller-runtime Manager
// This is a helper function that modifies ctrl.Options to enable leader election
func Setup(opts Options) func(*ctrl.Options) {
	return func(managerOpts *ctrl.Options) {
		if !opts.Enable {
			return // Leader election disabled, no changes needed
		}

		// Set leader election ID (Lease name)
		if opts.LeaseName != "" {
			managerOpts.LeaderElectionID = opts.LeaseName
		}

		// Enable leader election
		managerOpts.LeaderElection = true

		// Set namespace if provided
		if opts.Namespace != "" {
			managerOpts.LeaderElectionNamespace = opts.Namespace
		}

		// Set lease duration
		if opts.LeaseDuration > 0 {
			managerOpts.LeaseDuration = func() *time.Duration {
				d := opts.LeaseDuration
				return &d
			}()
		}

		// Set renew deadline
		if opts.RenewDeadline > 0 {
			managerOpts.RenewDeadline = func() *time.Duration {
				d := opts.RenewDeadline
				return &d
			}()
		}

		// Set retry period
		if opts.RetryPeriod > 0 {
			managerOpts.RetryPeriod = func() *time.Duration {
				d := opts.RetryPeriod
				return &d
			}()
		}
	}
}

// ManagerOptions returns ctrl.Options with leader election configured
// This is a convenience function for creating manager options directly
func ManagerOptions(baseOpts ctrl.Options, leaderOpts Options) ctrl.Options {
	// Apply leader election configuration
	setupFunc := Setup(leaderOpts)
	setupFunc(&baseOpts)

	return baseOpts
}

