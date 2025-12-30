# Migration Examples: zen-flow and zen-lock

This document shows practical examples of migrating zen-flow and zen-lock to use zen-sdk.

## zen-flow Migration

### Current Implementation

**File:** `zen-flow/pkg/controller/manager.go`

```go
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
```

### Migrated Implementation

**Step 1:** Add dependency

```bash
cd zen-flow
go get github.com/kube-zen/zen-sdk@latest
go mod tidy
```

**Step 2:** Update `zen-flow/pkg/controller/manager.go`

```go
import "github.com/kube-zen/zen-sdk/pkg/leader"

func ManagerOptions(namespace string, enableLeaderElection bool) ctrl.Options {
    baseOpts := ctrl.Options{
        Scheme: nil,
    }
    
    leaderOpts := leader.Options{
        LeaseName:  "zen-flow-controller-leader-election",
        Enable:     enableLeaderElection,
        Namespace:  namespace,
        // Uses defaults: 15s lease, 10s renew, 2s retry
    }
    
    return leader.ManagerOptions(baseOpts, leaderOpts)
}
```

**Or simplify in main.go:**

```go
import "github.com/kube-zen/zen-sdk/pkg/leader"

leaderOpts := leader.Options{
    LeaseName: "zen-flow-controller-leader-election",
    Enable:    enableLeaderElection,
    Namespace: namespace,
}

mgr, err := ctrl.NewManager(cfg, ctrl.Options{
    Scheme: scheme,
}, leader.Setup(leaderOpts))
```

**Step 3:** Remove `ManagerOptions` function (no longer needed)

**Step 4:** Add metrics (optional)

```go
import "github.com/kube-zen/zen-sdk/pkg/metrics"

recorder := metrics.NewRecorder("zen-flow")

// In reconciler
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
    start := time.Now()
    // ... reconciliation logic ...
    duration := time.Since(start).Seconds()
    r.recorder.RecordReconciliationSuccess(duration)
    return result, nil
}
```

### Benefits

- ✅ **10+ lines removed** - No custom leader election code
- ✅ **Consistent** - Same API as other tools
- ✅ **Maintainable** - Fix once in SDK, benefits all
- ✅ **Tested** - SDK is well-tested

---

## zen-lock Migration

### Current Implementation

**File:** `zen-lock/cmd/webhook/main.go`

```go
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
```

### Migrated Implementation

**Step 1:** Add dependency

```bash
cd zen-lock
go get github.com/kube-zen/zen-sdk@latest
go mod tidy
```

**Step 2:** Update `zen-lock/cmd/webhook/main.go`

```go
import (
    "github.com/kube-zen/zen-sdk/pkg/leader"
    "github.com/kube-zen/zen-sdk/pkg/logging"
    "github.com/kube-zen/zen-sdk/pkg/metrics"
    sdkwebhook "github.com/kube-zen/zen-sdk/pkg/webhook"
)

func main() {
    // Use zen-sdk logging
    logger := logging.NewLogger("zen-lock")
    
    // Setup metrics
    recorder := metrics.NewRecorder("zen-lock")
    
    // Configure leader election
    leaderOpts := leader.Options{
        LeaseName: "zen-lock-webhook-leader-election",
        Enable:    enableLeaderElection,
    }
    
    // Create manager with leader election
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
    
    // ... rest of setup ...
}
```

**Step 3:** Use webhook helpers (if applicable)

```go
// In webhook handler
import sdkwebhook "github.com/kube-zen/zen-sdk/pkg/webhook"

// Generate patch
updates := map[string]interface{}{
    "/metadata/labels/managed-by": "zen-lock",
}
patch, err := sdkwebhook.GeneratePatch(obj, updates)
return admission.Patched("", patch)
```

### Benefits

- ✅ **Consistent leader election** - Same as zen-flow
- ✅ **Structured logging** - Better observability
- ✅ **Standard metrics** - Consistent metrics
- ✅ **Webhook helpers** - Easier patch generation

---

## Migration Checklist

### For zen-flow

- [ ] Add `github.com/kube-zen/zen-sdk` dependency
- [ ] Replace `ManagerOptions` with `leader.Setup()`
- [ ] (Optional) Add `metrics.NewRecorder()`
- [ ] (Optional) Add `logging.NewLogger()`
- [ ] Remove custom leader election code
- [ ] Test leader election works
- [ ] Verify metrics (if added)
- [ ] Update documentation

### For zen-lock

- [ ] Add `github.com/kube-zen/zen-sdk` dependency
- [ ] Replace leader election config with `leader.Setup()`
- [ ] (Optional) Add `metrics.NewRecorder()`
- [ ] (Optional) Add `logging.NewLogger()`
- [ ] (Optional) Use `webhook.GeneratePatch()`
- [ ] Test leader election works
- [ ] Test webhook still works
- [ ] Verify metrics (if added)
- [ ] Update documentation

---

## Code Reduction Summary

### zen-flow

**Before:**
- `ManagerOptions` function: ~15 lines
- Custom leader election config: ~10 lines
- **Total: ~25 lines**

**After:**
- Import zen-sdk: 1 line
- Use `leader.Setup()`: 5 lines
- **Total: ~6 lines**

**Reduction: ~19 lines (76% reduction)**

### zen-lock

**Before:**
- Leader election in main.go: ~5 lines
- Custom metrics (if any): ~20 lines
- Custom logging (if any): ~15 lines
- **Total: ~40 lines**

**After:**
- Import zen-sdk: 3 lines
- Use SDK packages: ~10 lines
- **Total: ~13 lines**

**Reduction: ~27 lines (68% reduction)**

---

## Testing After Migration

1. **Leader Election:**
   ```bash
   # Deploy with 3 replicas
   kubectl scale deployment zen-flow --replicas=3
   
   # Check only 1 is leader
   kubectl get lease zen-flow-controller-leader-election
   ```

2. **Metrics:**
   ```bash
   # Check metrics endpoint
   curl http://localhost:8080/metrics | grep zen_
   ```

3. **Logging:**
   ```bash
   # Check logs have component name
   kubectl logs deployment/zen-flow | grep component
   ```

---

**See [examples/zen-flow-migration.go](examples/zen-flow-migration.go) and [examples/zen-lock-migration.go](examples/zen-lock-migration.go) for complete examples.**

