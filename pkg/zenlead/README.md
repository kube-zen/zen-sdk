# zenlead Package

This package provides a unified leadership configuration API for all Kube-ZEN components.

## Overview

The `zenlead` package implements the [Leadership Contract](../../docs/LEADERSHIP_CONTRACT.md) by providing:

1. **Unified configuration**: Single API for all leadership modes
2. **Safety guards**: Runtime validation to prevent unsafe HA configurations
3. **Consistent defaults**: Standardized REST client and leader election settings

## Usage

### Profile B: Built-in Leader Election (Default)

```go
import (
    "github.com/kube-zen/zen-sdk/pkg/leader"
    "github.com/kube-zen/zen-sdk/pkg/zenlead"
    ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
    // Get namespace
    namespace, err := leader.RequirePodNamespace()
    if err != nil {
        log.Fatal(err)
    }

    // Apply REST config defaults
    restConfig := ctrl.GetConfigOrDie()
    zenlead.ControllerRuntimeDefaults(restConfig)

    // Configure leader election (Profile B: Built-in)
    leConfig := zenlead.LeaderElectionConfig{
        Mode:       zenlead.BuiltIn,
        ElectionID: "my-controller-leader-election",
        Namespace:  namespace,
    }

    baseOpts := ctrl.Options{
        Scheme: scheme,
    }

    mgrOpts, err := zenlead.PrepareManagerOptions(baseOpts, leConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Enforce safe HA (fail if replicas > 1 without leader election)
    replicaCount := 2 // from deployment
    if err := zenlead.EnforceSafeHA(replicaCount, mgrOpts.LeaderElection); err != nil {
        log.Fatal(err)
    }

    // Create manager
    mgr, err := ctrl.NewManager(restConfig, mgrOpts)
    // ...
}
```

### Profile C: Zen-lead Managed Leader Election

```go
    leConfig := zenlead.LeaderElectionConfig{
        Mode:       zenlead.ZenLeadManaged,
        LeaseName:  "my-controller-leader-group", // LeaderGroup CRD name
        Namespace:  namespace,
    }

    mgrOpts, err := zenlead.PrepareManagerOptions(baseOpts, leConfig)
    // ... rest is the same
```

### Profile A: Disabled (Single Replica Only)

```go
    leConfig := zenlead.LeaderElectionConfig{
        Mode: zenlead.Disabled,
    }

    mgrOpts, err := zenlead.PrepareManagerOptions(baseOpts, leConfig)
    
    // MUST enforce safety - will fail if replicas > 1
    replicaCount := 1
    if err := zenlead.EnforceSafeHA(replicaCount, mgrOpts.LeaderElection); err != nil {
        log.Fatal(err) // Will fail if replicas > 1
    }
```

## API Reference

### Types

- `LeadershipMode`: `builtin` | `zenlead` | `disabled`
- `LeaderElectionConfig`: Configuration struct

### Functions

- `ControllerRuntimeDefaults(cfg *rest.Config)`: Apply REST client defaults
- `PrepareManagerOptions(base ctrl.Options, le LeaderElectionConfig) (ctrl.Options, error)`: Configure leader election
- `EnforceSafeHA(replicaCount int, leaderElectionEnabled bool) error`: Validate safe HA configuration

## Safety Guarantees

1. **Unsafe configurations fail fast**: `EnforceSafeHA()` hard-fails if `replicas > 1` without leader election
2. **Validation**: `PrepareManagerOptions()` validates required fields per mode
3. **Deterministic naming**: ZenLeadManaged mode derives ElectionID from LeaseName consistently

## See Also

- [Leadership Contract](../../docs/LEADERSHIP_CONTRACT.md) - Full contract definition
- [zen-lead README](../../../zen-lead/README.md) - Network-only profile (Profile A)

