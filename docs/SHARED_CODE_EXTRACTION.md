# Shared-Code Extraction Operating Model

**Version**: 1.0.0  
**Last Updated**: 2025-01-15  
**Purpose**: Formalize when and how code gets promoted into zen-sdk

---

## Promotion Criteria

Code MUST meet ALL of the following criteria before being promoted to zen-sdk:

### 1. Multi-Component Usage

- ✅ **Used by ≥2 components** OR
- ✅ **Clearly intended as a platform primitive** (e.g., leadership, metrics, logging)

**Rationale**: Prevents premature abstraction. Single-component code stays in that component.

### 2. Stable API Surface

- ✅ **Versioned package** (e.g., `pkg/gc/v1alpha1`)
- ✅ **No breaking churn** (API changes require version bump)
- ✅ **Clear interface boundaries** (not implementation details)

**Rationale**: zen-sdk is a dependency; breaking changes affect all consumers.

### 3. Unit Tests in zen-sdk

- ✅ **Comprehensive unit tests** (≥80% coverage)
- ✅ **Test harness with fake client** where possible
- ✅ **Integration test examples** (if applicable)

**Rationale**: Shared code must be well-tested since it affects multiple components.

### 4. No Circular Dependencies

- ✅ **Components depend on zen-sdk** (one-way dependency)
- ✅ **zen-sdk never depends on components**
- ✅ **No component-specific logic in zen-sdk**

**Rationale**: Circular dependencies create build and versioning nightmares.

---

## Extraction Process

### Step 1: Identify Duplication/Shared Behavior

**Signals**:
- Same logic appears in ≥2 components
- Similar patterns with minor variations
- Common utilities (rate limiting, backoff, selectors, etc.)

**Tools**:
- Code search across repos
- Pattern analysis (grep for similar function names)
- Architecture review

### Step 2: Carve Out Minimal Library API

**Location**: `zen-sdk/pkg/<capability>/`

**Structure**:
```
zen-sdk/pkg/<capability>/
  ├── v1alpha1/          # Versioned API (if needed)
  │   ├── types.go       # Core types
  │   └── interface.go   # Public interfaces
  ├── collector.go       # Main implementation
  ├── collector_test.go  # Unit tests
  └── README.md          # Usage documentation
```

**Principles**:
- **Minimal API**: Only expose what's needed
- **Interface-based**: Allow different implementations
- **Composable**: Build complex behavior from simple primitives

### Step 3: Migrate Callers

**Process**:
1. Create zen-sdk package with tests
2. Update component `go.mod` to use zen-sdk (tagged version)
3. Replace component code with zen-sdk import
4. Verify compilation and tests pass
5. Remove duplicated code from component

**Migration Pattern**:
```go
// Before (in component)
import "github.com/kube-zen/zen-gc/pkg/controller/rate_limiter"

// After (using zen-sdk)
import "github.com/kube-zen/zen-sdk/pkg/gc/ratelimiter"
```

### Step 4: Delete Duplicated Code

**After migration**:
- ✅ Remove duplicated files from components
- ✅ Update component documentation
- ✅ Verify no remaining references

### Step 5: Add CI Guardrails

**Prevent re-divergence**:
- ✅ CI check for banned package paths (H114)
- ✅ Denylist for component-specific implementations of shared capabilities
- ✅ Documentation of promoted capabilities

---

## Example: GC Capability Extraction (H112)

### Identified Shared Capabilities

1. **Rate Limiting**
   - zen-gc: `pkg/controller/rate_limiter.go`
   - zen-watcher: `pkg/server/ratelimit.go`
   - **Shared**: Token bucket rate limiting

2. **TTL Evaluation**
   - zen-gc: `pkg/controller/ttl_test.go`, `evaluate_policy_shared.go`
   - zen-watcher: `pkg/gc/collector.go` (age-based deletion)
   - **Shared**: Time-based expiration logic

3. **Field Path Evaluation**
   - zen-gc: `pkg/controller/field_path.go`
   - **Shared**: Extract values from resource fields (JSONPath-like)

4. **Selector Matching**
   - zen-gc: `pkg/controller/selectors_test.go`
   - **Shared**: Label/field selector evaluation

5. **Event Recording**
   - zen-gc: `pkg/controller/events.go`
   - **Shared**: Kubernetes event emission patterns

6. **Backoff/Retry Logic**
   - zen-gc: `pkg/controller/backoff.go`
   - **Shared**: Exponential backoff for retries

### Extraction Plan

**Phase 1: Core Primitives**
- `zen-sdk/pkg/gc/ratelimiter` - Rate limiting
- `zen-sdk/pkg/gc/backoff` - Backoff/retry logic
- `zen-sdk/pkg/gc/fieldpath` - Field path evaluation

**Phase 2: GC-Specific Logic**
- `zen-sdk/pkg/gc/collector` - GC collection loop
- `zen-sdk/pkg/gc/selector` - Resource selector matching
- `zen-sdk/pkg/gc/ttl` - TTL evaluation

**Phase 3: Integration**
- `zen-sdk/pkg/gc/events` - Event recording
- `zen-sdk/pkg/gc/metrics` - Metrics hooks

---

## CI Guardrails (H114)

### Banned Package Paths

After extraction, component repos MUST NOT recreate:

- `internal/gc/*` - GC logic must live in zen-sdk/pkg/gc
- `pkg/ratelimiter/*` - Rate limiting must use zen-sdk/pkg/gc/ratelimiter
- `pkg/backoff/*` - Backoff must use zen-sdk/pkg/gc/backoff

### CI Check

**Script**: `scripts/ci/check-banned-packages.sh`

**Functionality**:
- Scans for banned package paths
- Fails if shared capability is re-implemented
- Provides migration guidance

---

## Versioning Strategy

### zen-sdk Versioning

- **v0.x.x-alpha**: Pre-1.0, may have breaking changes
- **v1.x.x**: Stable API, backward-compatible changes only
- **v2.x.x**: Breaking changes require major version bump

### Component Dependency Pinning

- **Always pin to tagged version** (never pseudo-versions)
- **Update go.mod** when zen-sdk releases new version
- **Test against tagged version** before upgrading

---

## Exit Criteria

✅ "Move to zen-sdk" is a predictable playbook, not ad-hoc  
✅ Promotion criteria are clear and enforced  
✅ Extraction process is documented and repeatable  
✅ CI guardrails prevent re-divergence

---

## Related

- [Leadership Contract](../LEADERSHIP_CONTRACT.md) - Contract for leadership code
- [H112 GC Extraction Plan](./GC_EXTRACTION_PLAN.md) - Specific GC extraction plan
- [H114 CI Guardrails](./CI_GUARDRAILS.md) - Banned package enforcement

