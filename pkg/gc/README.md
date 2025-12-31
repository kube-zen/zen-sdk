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

## Extraction Status

- ✅ **ratelimiter**: Extracted (H112)
- ✅ **backoff**: Extracted (H112)
- ⏳ **fieldpath**: Planned
- ⏳ **ttl**: Planned
- ⏳ **selector**: Planned

## Migration Guide

See [GC Extraction Plan](../docs/GC_EXTRACTION_PLAN.md) for detailed migration steps.

