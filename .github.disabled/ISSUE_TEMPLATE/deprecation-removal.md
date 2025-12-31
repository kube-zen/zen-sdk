# Deprecation Removal: zen-sdk/pkg/controller

**Target Version**: v1.0.0  
**Status**: Tracked for removal  
**Created**: 2025-01-15

## Overview

The `zen-sdk/pkg/controller` package contains deprecated leader guard code that relies on pod annotations (`zen-lead/role`), which violates the leadership contract's "no pod mutation" Day-0 guarantee.

## Removal Timeline

- **v0.1.0-alpha** (current): Package quarantined, excluded from denylist scans
- **v1.0.0** (target): Package will be removed entirely

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

See `pkg/controller/REMOVAL_NOTICE.md` for detailed migration guide.

## CI Guard

A CI guard prevents new imports of `zen-sdk/pkg/controller` outside the quarantined boundary:

```bash
# Script: scripts/ci/check-controller-imports.sh
# Fails if pkg/controller is imported in non-quarantined code
```

## Acceptance Criteria

- [ ] All components migrated to `zen-sdk/pkg/zenlead`
- [ ] No active code uses `pkg/controller`
- [ ] Migration guide published
- [ ] Removal notice updated with final removal date
- [ ] Package deleted from repository

## Related

- `pkg/controller/REMOVAL_NOTICE.md` - Migration guide
- `docs/LEADERSHIP_CONTRACT.md` - Contract definition
- H090 - Initial quarantine

