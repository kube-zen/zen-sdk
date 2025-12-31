# GC Capability Consolidation Plan (H112)

**Status**: Planning  
**Date**: 2025-01-15  
**Target**: Extract reusable GC logic from zen-gc and zen-watcher into zen-sdk

---

## Inventory: GC Code in Context

### zen-gc Capabilities

1. **Rate Limiting** (`pkg/controller/rate_limiter.go`)
   - Token bucket algorithm (golang.org/x/time/rate)
   - Per-policy rate limiters
   - Dynamic rate updates

2. **Backoff/Retry Logic** (`pkg/controller/backoff.go`)
   - Exponential backoff for deletions
   - Configurable steps, duration, factor, jitter, cap

3. **TTL Evaluation** (`pkg/controller/ttl_test.go`, `evaluate_policy_shared.go`)
   - Fixed TTL (seconds after creation)
   - Field-based TTL (extract from resource)
   - Mapped TTL (different TTLs by value)
   - Relative TTL (relative to timestamp field)

4. **Field Path Evaluation** (`pkg/controller/field_path.go`)
   - JSONPath-like field extraction
   - Type conversion and validation

5. **Selector Matching** (`pkg/controller/selectors_test.go`)
   - Label selector evaluation
   - Field selector evaluation

6. **Event Recording** (`pkg/controller/events.go`)
   - Kubernetes event emission
   - Event categorization

### zen-watcher Capabilities

1. **GC Collector** (`pkg/gc/collector.go`)
   - Age-based deletion (TTL in days)
   - Chunked listing (prevents memory issues)
   - Namespace scanning
   - Metrics integration

2. **Rate Limiting** (`pkg/server/ratelimit.go`)
   - Custom token bucket implementation
   - Per-key rate limiting (HTTP client IP)
   - Cleanup of old entries

---

## Shared Capabilities to Extract

### Phase 1: Core Primitives (High Reuse)

#### 1. Rate Limiter (`zen-sdk/pkg/gc/ratelimiter`)

**API**:
```go
package ratelimiter

type RateLimiter interface {
    Wait(ctx context.Context) error
    Allow() bool
    SetRate(maxPerSecond int)
}

// NewRateLimiter creates a rate limiter using token bucket algorithm
func NewRateLimiter(maxPerSecond int) RateLimiter
```

**Rationale**: Both zen-gc and zen-watcher need rate limiting. zen-gc uses golang.org/x/time/rate, zen-watcher has custom implementation. Standardize on one.

#### 2. Backoff (`zen-sdk/pkg/gc/backoff`)

**API**:
```go
package backoff

type Backoff struct {
    Steps    int
    Duration time.Duration
    Factor   float64
    Jitter   float64
    Cap      time.Duration
}

func (b *Backoff) Next() time.Duration
func (b *Backoff) Reset()
```

**Rationale**: zen-gc has backoff logic that could be reused by zen-watcher for retries.

#### 3. Field Path (`zen-sdk/pkg/gc/fieldpath`)

**API**:
```go
package fieldpath

func ExtractValue(obj *unstructured.Unstructured, path string) (interface{}, error)
func ExtractString(obj *unstructured.Unstructured, path string) (string, error)
func ExtractInt64(obj *unstructured.Unstructured, path string) (int64, error)
```

**Rationale**: Field path evaluation is a common pattern for extracting values from resources.

### Phase 2: GC-Specific Logic (Medium Reuse)

#### 4. TTL Evaluator (`zen-sdk/pkg/gc/ttl`)

**API**:
```go
package ttl

type TTLSpec struct {
    SecondsAfterCreation *int64
    FieldPath            string
    Mappings             map[string]int64
    Default              *int64
    RelativeTo           string
    SecondsAfter         *int64
}

type Evaluator interface {
    CalculateTTL(obj *unstructured.Unstructured, spec *TTLSpec) (int64, error)
    ShouldDelete(obj *unstructured.Unstructured, spec *TTLSpec) (bool, error)
}

func NewEvaluator() Evaluator
```

**Rationale**: TTL evaluation logic is complex and reusable. zen-watcher's age-based deletion is a simpler case.

#### 5. Selector Matcher (`zen-sdk/pkg/gc/selector`)

**API**:
```go
package selector

func MatchesLabelSelector(obj *unstructured.Unstructured, selector metav1.LabelSelector) (bool, error)
func MatchesFieldSelector(obj *unstructured.Unstructured, selector fields.Selector) (bool, error)
```

**Rationale**: Selector matching is a common Kubernetes pattern.

### Phase 3: Integration (Lower Priority)

#### 6. Event Recorder (`zen-sdk/pkg/gc/events`)

**API**:
```go
package events

type Recorder interface {
    RecordDeletion(obj *unstructured.Unstructured, reason string)
    RecordError(obj *unstructured.Unstructured, err error)
}
```

**Rationale**: Event recording patterns can be standardized.

---

## Implementation Plan

### Step 1: Create zen-sdk/pkg/gc Structure

```
zen-sdk/pkg/gc/
  ├── ratelimiter/
  │   ├── ratelimiter.go
  │   ├── ratelimiter_test.go
  │   └── README.md
  ├── backoff/
  │   ├── backoff.go
  │   ├── backoff_test.go
  │   └── README.md
  ├── fieldpath/
  │   ├── fieldpath.go
  │   ├── fieldpath_test.go
  │   └── README.md
  ├── ttl/
  │   ├── evaluator.go
  │   ├── evaluator_test.go
  │   └── README.md
  └── README.md
```

### Step 2: Migrate zen-gc

1. Update `zen-gc/go.mod` to use zen-sdk v0.1.1-alpha (after GC extraction)
2. Replace `pkg/controller/rate_limiter.go` with `zen-sdk/pkg/gc/ratelimiter`
3. Replace `pkg/controller/backoff.go` with `zen-sdk/pkg/gc/backoff`
4. Replace `pkg/controller/field_path.go` with `zen-sdk/pkg/gc/fieldpath`
5. Replace TTL evaluation with `zen-sdk/pkg/gc/ttl`
6. Remove duplicated code
7. Run tests to verify

### Step 3: Migrate zen-watcher

1. Update `zen-watcher/go.mod` to use zen-sdk v0.1.1-alpha
2. Replace `pkg/server/ratelimit.go` with `zen-sdk/pkg/gc/ratelimiter` (HTTP middleware can wrap it)
3. Replace `pkg/gc/collector.go` TTL logic with `zen-sdk/pkg/gc/ttl`
4. Keep collector loop (component-specific) but use shared TTL evaluator
5. Remove duplicated code
6. Run tests to verify

### Step 4: Add Tests

- Unit tests for each package (≥80% coverage)
- Integration tests with fake client
- Example usage in README

---

## Exit Criteria

✅ One GC logic implementation in zen-sdk  
✅ Both zen-gc and zen-watcher consume zen-sdk/pkg/gc  
✅ Components become thin wiring layers  
✅ No duplicated GC code in components

---

## Notes

- **Incremental Approach**: Extract one capability at a time (ratelimiter → backoff → fieldpath → ttl)
- **Backward Compatibility**: Maintain existing component APIs during migration
- **Testing**: Each extraction must pass all component tests before proceeding

