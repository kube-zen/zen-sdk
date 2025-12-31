# GC Package

Shared garbage collection primitives extracted from zen-gc and zen-watcher.

## Packages

### ratelimiter

Rate limiting using token bucket algorithm.

```go
import "github.com/kube-zen/zen-sdk/pkg/gc/ratelimiter"

rl := ratelimiter.NewRateLimiter(10) // 10 ops/sec
if err := rl.Wait(ctx); err != nil {
    return err
}
```

### backoff

Exponential backoff for retry operations.

```go
import "github.com/kube-zen/zen-sdk/pkg/gc/backoff"

b := backoff.NewBackoff(backoff.DefaultConfig())
for !b.IsExhausted() {
    duration := b.Next()
    time.Sleep(duration)
    // retry operation
}
```

### ttl

TTL (Time-To-Live) evaluation for resource expiration.

```go
import "github.com/kube-zen/zen-sdk/pkg/gc/ttl"

// Fixed TTL: 1 hour after creation
ttlSeconds := int64(3600)
spec := &ttl.Spec{
    SecondsAfterCreation: &ttlSeconds,
}

expired, err := ttl.IsExpired(resource, spec)
if err == nil && expired {
    // Delete the resource
}
```

### fieldpath

Field path evaluation for extracting values from resources.

```go
import "github.com/kube-zen/zen-sdk/pkg/gc/fieldpath"

severity, found, _ := fieldpath.GetString(resource, "spec.severity")
if found && severity == "critical" {
    // Handle critical resource
}
```

### selector

Resource selector matching (labels, annotations, fields, phase).

```go
import "github.com/kube-zen/zen-sdk/pkg/gc/selector"

conditions := &selector.Conditions{
    Phase: []string{"Failed", "Succeeded"},
    HasLabels: []selector.LabelCondition{
        {Key: "app", Value: "myapp"},
    },
}

if selector.MatchesConditions(resource, conditions) {
    // Resource matches all conditions
}
```

## Extraction Status

- ✅ **ratelimiter**: Extracted (H112 Phase 1)
- ✅ **backoff**: Extracted (H112 Phase 1)
- ✅ **ttl**: Extracted (H112 Phase 2)
- ✅ **fieldpath**: Extracted (H112 Phase 2)
- ✅ **selector**: Extracted (H112 Phase 2)

## Migration Guide

See [GC Extraction Plan](../../docs/GC_EXTRACTION_PLAN.md) for detailed migration steps.

