# Leadership Contract

**Version**: 1.0.0  
**Last Updated**: 2025-01-15  
**Stability**: Stable (v1.0.0+)

## Change Process

- **Major version bump**: Breaking changes to API or behavior
- **Minor version bump**: New features, backward-compatible
- **Patch version bump**: Bug fixes, documentation updates

## Version History

- **v1.0.0** (2025-01-15): Initial stable release
  - Standardized leadership profiles (A, B, C)
  - CI denylist enforcement
  - Helm chart safety guards

---

This document defines the authoritative leadership model for all Kube-ZEN components. All components MUST follow one of three supported profiles.

## Leadership Profiles

### Profile A — Network-only (Zero Code Changes)

**Purpose**: Route client traffic to a single active pod without application code changes.

**How it works**:
- Annotate any Service with `zen-lead.io/enabled: "true"`
- zen-lead creates a selector-less leader Service (`<svc-name>-leader`)
- zen-lead manages an EndpointSlice pointing to exactly one leader pod
- Applications connect to the leader Service instead of the original Service
- No pod mutation, no CRDs required

**Use case**: Any workload that needs single-active routing (databases, stateful services, etc.)

**Day-0 support**: ✅ Fully supported without CRDs

---

### Profile B — Controller HA (Built-in, K8s-native)

**Purpose**: Ensure only one controller replica actively reconciles resources.

**How it works**:
- Component uses Kubernetes Lease API via controller-runtime
- Configured via `zen-sdk/pkg/zenlead.PrepareManagerOptions()` with `mode: builtin`
- Leader election is Lease-based (etcd-backed)
- Only the leader pod runs reconcilers
- Follower pods wait for leadership

**Use case**: Controllers that need HA but don't need network-level routing

**Day-0 support**: ✅ Fully supported (standard Kubernetes Lease API)

---

### Profile C — Zen-lead Managed Leadership (CRD-driven)

**Purpose**: zen-lead provisions and owns the Lease; components consume it with minimal code changes.

**How it works**:
- Create a `LeaderGroup` CRD with `spec.type: controller`
- zen-lead ensures a Lease exists with deterministic name/labels
- Component uses `zen-sdk/pkg/zenlead.PrepareManagerOptions()` with `mode: zenlead`
- Component consumes the Lease (same Lease API as Profile B)
- zen-lead updates LeaderGroup status from Lease

**Use case**: Centralized leadership management, advanced fencing, or when zen-lead is already deployed

**Day-0 support**: ⚠️ Requires CRDs (optional module, does not break Profile A)

---

## Source of Truth Objects

### Lease (The Lock)
- **Purpose**: etcd-backed coordination primitive
- **Used by**: Profile B (directly), Profile C (via zen-lead)
- **API**: `coordination.k8s.io/v1/Lease`
- **Ownership**: Component (Profile B) or zen-lead (Profile C)

### CRD (Intent/Status Projection)
- **Purpose**: Declarative leadership intent and status
- **Used by**: Profile C only
- **API**: `leadership.kube-zen.io/v1alpha1/LeaderGroup`
- **Ownership**: zen-lead

### EndpointSlice (Routing)
- **Purpose**: Network-level routing to leader pod
- **Used by**: Profile A only
- **API**: `discovery.k8s.io/v1/EndpointSlice`
- **Ownership**: zen-lead

---

## Prohibited Patterns

The following patterns are **FORBIDDEN** and MUST NOT be used:

1. **Pod annotation watcher**: Components MUST NOT watch pod annotations for leadership status
   - ❌ `zen-lead/role` annotation
   - ❌ `NewWatcher()` pattern
   - ❌ External watcher polling

2. **"zen-lead for controller HA via network routing"**: Controllers MUST NOT use Profile A for HA
   - ❌ "Use zen-lead Service routing for controller HA"
   - ❌ Controllers connecting to leader Service for coordination
   - ✅ Controllers use Profile B (built-in Lease) or Profile C (zen-lead managed Lease)

3. **External HA modes**: No "external" or "zen-lead managed" modes that mutate pods
   - ❌ `ha-mode=external`
   - ❌ `ha-mode=zenlead`
   - ✅ Use Profile B or Profile C via `zen-sdk/pkg/zenlead`

---

## Implementation Requirements

### For Component Authors

1. **Use zen-sdk wrapper**: All components MUST use `zen-sdk/pkg/zenlead.PrepareManagerOptions()`
2. **Support Profile B by default**: Default to `mode: builtin` (Profile B)
3. **Optional Profile C**: Support `mode: zenlead` if zen-lead is deployed
4. **Safety guards**: Call `EnforceSafeHA()` at startup to prevent unsafe configurations

### For zen-lead Authors

1. **Profile A is default**: Network-only mode MUST work without CRDs
2. **Profile C is optional**: CRD module MUST be opt-in (`--set leadergroups.enabled=true`)
3. **No pod mutation**: zen-lead MUST NOT mutate workload pods (Day-0 contract)

---

## CI Denylist (H078: Runtime-only Enforcement)

All repositories MUST enforce the following denylist in CI using **runtime-only** enforcement scope.

### Enforcement Scope

**Included** (scanned for violations):
- `/cmd` - Application entry points
- `/pkg` - Public packages
- `/internal` - Internal packages
- `/charts/templates` - Helm chart templates

**Excluded** (not scanned):
- `/docs` - Documentation files
- `*.md` - All markdown files (may reference banned patterns for explanation)
- `/test`, `/tests`, `/e2e` - Test files (may reference banned patterns for testing)
- `CHANGELOG.md` - Release notes
- Deprecated code marked with `DEPRECATED` comments

### Rationale

Runtime-only enforcement ensures:
1. **Actual code violations are caught** - Banned patterns in executable code are detected
2. **Documentation remains flexible** - Docs can explain prohibited patterns without triggering false positives
3. **Test code is exempt** - Tests may legitimately reference banned patterns to verify they're not used
4. **Deprecated code is handled** - Legacy code marked as deprecated is excluded

### Implementation

The denylist CI check (`scripts/ci/validate-leadership-denylist.sh`) scans runtime code paths only:

```bash
# Prohibited patterns (must return 0 matches in runtime code)
# Scanned paths: /cmd, /pkg, /internal, /charts/templates
# Excluded: /docs, *.md, test files, deprecated code
```

**Banned patterns**:
- `NewWatcher` - Pod annotation watcher pattern
- `zen-lead/role` - Pod role annotation
- `ha-mode=external` - External HA mode
- `use zen-lead for controller HA` - Misuse of zen-lead for controller HA

If any matches are found in runtime code, CI MUST fail.

---

## Migration Guide

### From Old "External Watcher" Pattern

**Old (FORBIDDEN)**:
```go
watcher := leader.NewWatcher(...)
if watcher.IsLeader() {
    // reconcile
}
```

**New (Profile B)**:
```go
leConfig := zenlead.LeaderElectionConfig{
    Mode: zenlead.BuiltIn,
    ElectionID: "my-controller-leader-election",
    Namespace: namespace,
}
opts, err := zenlead.PrepareManagerOptions(baseOpts, leConfig)
```

**New (Profile C)**:
```go
leConfig := zenlead.LeaderElectionConfig{
    Mode: zenlead.ZenLeadManaged,
    LeaseName: "my-controller-leader-group",
    Namespace: namespace,
}
opts, err := zenlead.PrepareManagerOptions(baseOpts, leConfig)
```

---

## References

- [zen-sdk/pkg/zenlead](../pkg/zenlead/README.md) - Implementation details
- [zen-lead README](../../zen-lead/README.md) - Network-only profile (Profile A)
- [Kubernetes Lease API](https://kubernetes.io/docs/concepts/architecture/leases/) - Profile B/C foundation

