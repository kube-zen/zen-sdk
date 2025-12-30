# Zen SDK Quick Start

Get started with zen-sdk in 5 minutes!

## Installation

```bash
go get github.com/kube-zen/zen-sdk@latest
```

## Usage Examples

### 1. Leader Election

Enable HA leader election in your controller:

```go
import (
    "github.com/kube-zen/zen-sdk/pkg/leader"
    ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
    opts := leader.Options{
        LeaseName: "my-controller-leader-election",
        Enable:    true,
        Namespace: "default",
    }
    
    mgr, err := ctrl.NewManager(cfg, ctrl.Options{
        Scheme: scheme,
    }, leader.Setup(opts))
    
    // Manager now has leader election enabled
}
```

### 2. Metrics

Record metrics in your reconciler:

```go
import "github.com/kube-zen/zen-sdk/pkg/metrics"

// In controller setup
recorder := metrics.NewRecorder("my-controller")

// In reconciler
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
    start := time.Now()
    
    // ... your logic ...
    
    duration := time.Since(start).Seconds()
    if err != nil {
        r.recorder.RecordReconciliationError(duration)
    } else {
        r.recorder.RecordReconciliationSuccess(duration)
    }
    
    return result, err
}
```

### 3. Logging

Use structured logging:

```go
import "github.com/kube-zen/zen-sdk/pkg/logging"

func main() {
    logger := logging.NewLogger("my-controller")
    logger.Info("Controller started")
    
    logger.WithField("namespace", "default").Info("Processing namespace")
}
```

### 4. Webhook Helpers

Generate JSON patches:

```go
import "github.com/kube-zen/zen-sdk/pkg/webhook"

// Add a label
patch, err := webhook.GenerateAddPatch("/metadata/labels/managed-by", "zen-sdk")

// Remove a label
patch, err := webhook.GenerateRemovePatch("/metadata/labels/test")

// Generate patch from updates
updates := map[string]interface{}{
    "/metadata/labels/test": "value",
}
patch, err := webhook.GeneratePatch(obj, updates)
```

## Complete Example

Here's a complete controller setup using all SDK packages:

```go
package main

import (
    "context"
    "time"
    
    "github.com/kube-zen/zen-sdk/pkg/leader"
    "github.com/kube-zen/zen-sdk/pkg/logging"
    "github.com/kube-zen/zen-sdk/pkg/metrics"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func main() {
    // Setup logging
    logger := logging.NewLogger("my-controller")
    logger.Info("Starting controller")
    
    // Setup metrics
    recorder := metrics.NewRecorder("my-controller")
    
    // Setup leader election
    opts := leader.Options{
        LeaseName: "my-controller-leader-election",
        Enable:    true,
    }
    
    // Create manager
    mgr, err := ctrl.NewManager(cfg, ctrl.Options{
        Scheme: scheme,
    }, leader.Setup(opts))
    if err != nil {
        logger.Error(err, "Failed to create manager")
        os.Exit(1)
    }
    
    // Setup reconciler
    reconciler := &Reconciler{
        Client:   mgr.GetClient(),
        recorder: recorder,
        logger:   logger,
    }
    
    // Start manager
    logger.Info("Starting manager")
    if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
        logger.Error(err, "Problem running manager")
        os.Exit(1)
    }
}

type Reconciler struct {
    client.Client
    recorder *metrics.Recorder
    logger   *logging.Logger
}

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
    start := time.Now()
    
    r.logger.WithField("namespace", req.Namespace).
        WithField("name", req.Name).
        Info("Reconciling")
    
    // ... your reconciliation logic ...
    
    duration := time.Since(start).Seconds()
    r.recorder.RecordReconciliationSuccess(duration)
    
    return reconcile.Result{}, nil
}
```

## Next Steps

- Read [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) to migrate existing tools
- Check [ARCHITECTURE.md](ARCHITECTURE.md) for design details
- See [examples/](examples/) for more examples

---

**That's it! You're ready to use zen-sdk.** 🚀

