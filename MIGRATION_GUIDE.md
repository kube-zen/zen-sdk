# Migration Guide: Using Zen SDK in Existing Tools

This guide shows how to migrate zen-flow, zen-lock, and other tools to use zen-sdk.

## Overview

**Goal**: Replace duplicate code with zen-sdk imports.

**Benefit**: Single source of truth, easier maintenance, consistent behavior.

## Step 1: Add Dependency

```bash
cd zen-flow  # or zen-lock, zen-watcher, etc.
go get github.com/kube-zen/zen-sdk@latest
go mod tidy
```

## Step 2: Migrate Leader Election

### Before (zen-flow/pkg/controller/manager.go)

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

### After (Using zen-sdk)

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

**Or even simpler**:

```go
mgr, err := ctrl.NewManager(cfg, ctrl.Options{
    Scheme: scheme,
}, leader.Setup(leader.Options{
    LeaseName: "zen-flow-controller-leader-election",
    Enable:    enableLeaderElection,
    Namespace: namespace,
}))
```

## Step 3: Migrate Metrics

### Before (zen-flow/pkg/controller/metrics.go)

```go
// Custom metrics code (50+ lines)
var reconciliationsTotal = prometheus.NewCounterVec(...)
var reconciliationsDuration = prometheus.NewHistogramVec(...)
// ... registration code ...
```

### After (Using zen-sdk)

```go
import "github.com/kube-zen/zen-sdk/pkg/metrics"

// In controller setup
recorder := metrics.NewRecorder("zen-flow")

// In reconciler
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
    start := time.Now()
    
    // ... reconciliation logic ...
    
    duration := time.Since(start).Seconds()
    if err != nil {
        r.recorder.RecordReconciliationError(duration)
    } else {
        r.recorder.RecordReconciliationSuccess(duration)
    }
    
    return result, err
}
```

## Step 4: Migrate Logging

### Before (zen-flow/pkg/logging/logger.go)

```go
// Custom logging setup (30+ lines)
logger := zap.New(...)
// ... configuration ...
```

### After (Using zen-sdk)

```go
import "github.com/kube-zen/zen-sdk/pkg/logging"

logger := logging.NewLogger("zen-flow")
logger.Info("Controller started")
logger.WithField("namespace", "default").Info("Processing")
```

## Step 5: Migrate Webhook Helpers (zen-lock)

### Before (zen-lock/pkg/webhook/patch.go)

```go
// Custom patch generation (40+ lines)
func GeneratePatch(...) ([]byte, error) {
    // ... custom implementation ...
}
```

### After (Using zen-sdk)

```go
import "github.com/kube-zen/zen-sdk/pkg/webhook"

patch, err := webhook.GeneratePatch(obj, updates)
// or
patch, err := webhook.GenerateAddPatch("/metadata/labels/test", "value")
patch, err := webhook.GenerateRemovePatch("/metadata/labels/test")
```

## Step 6: Remove Duplicate Code

After migration, remove:

1. **Custom leader election code**
   - `pkg/controller/leader_election.go` (if exists)
   - Custom `ManagerOptions` implementation

2. **Custom metrics code**
   - `pkg/metrics/` directory (if exists)
   - Custom Prometheus setup

3. **Custom logging code**
   - `pkg/logging/` directory (if exists)
   - Custom zap configuration

4. **Custom webhook helpers**
   - `pkg/webhook/patch.go` (if exists)
   - Duplicate patch generation

## Step 7: Update Tests

Update tests to use zen-sdk:

```go
// Before
import "github.com/kube-zen/zen-flow/pkg/metrics"

// After
import "github.com/kube-zen/zen-sdk/pkg/metrics"
```

## Step 8: Verify

1. **Run tests:**
   ```bash
   go test ./...
   ```

2. **Build:**
   ```bash
   go build ./cmd/...
   ```

3. **Check metrics:**
   ```bash
   curl http://localhost:8080/metrics | grep zen_
   ```

## Migration Checklist

- [ ] Add zen-sdk dependency
- [ ] Replace leader election code
- [ ] Replace metrics code
- [ ] Replace logging code
- [ ] Replace webhook helpers (if applicable)
- [ ] Remove duplicate code
- [ ] Update tests
- [ ] Verify build
- [ ] Verify tests pass
- [ ] Update documentation

## Rollback Plan

If issues occur:

1. **Revert dependency:**
   ```bash
   go mod edit -droprequire github.com/kube-zen/zen-sdk
   go mod tidy
   ```

2. **Restore original code** from git history

3. **Investigate issue** and fix before retrying

## Benefits After Migration

✅ **Less code**: 150 lines → 50 lines (3x reduction)  
✅ **Consistent**: Same behavior across all tools  
✅ **Maintainable**: Fix once, benefits all tools  
✅ **Tested**: SDK is well-tested  
✅ **Documented**: Clear API and examples  

---

**Need help?** Check examples in `zen-sdk/examples/` or open an issue.

