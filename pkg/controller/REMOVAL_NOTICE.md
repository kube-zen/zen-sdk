# Removal Notice: pkg/controller Package

**Status**: Deprecated - Scheduled for removal in v1.0.0  
**Date**: 2025-01-15

## Overview

The `pkg/controller` package contains deprecated leader guard code that relies on pod annotations (`zen-lead/role`), which violates the leadership contract's "no pod mutation" Day-0 guarantee.

## Migration Path

**Use `zen-sdk/pkg/zenlead` instead:**

```go
// ❌ Old (deprecated)
import "github.com/kube-zen/zen-sdk/pkg/controller"
guard := controller.NewLeaderGuard(client, logger)
reconciler := guard.Wrap(innerReconciler)

// ✅ New (recommended)
import "github.com/kube-zen/zen-sdk/pkg/zenlead"
opts, err := zenlead.PrepareManagerOptions(baseOpts, leConfig)
mgr, err := ctrl.NewManager(cfg, opts)
```

## Removal Timeline

- **v0.1.0** (current): Package marked as deprecated, excluded from denylist scans
- **v1.0.0** (future): Package will be removed entirely

## Impact

No active code uses this package. It is kept temporarily for reference only and is excluded from CI denylist enforcement (Model A: runtime-only with legacy exclusion).

