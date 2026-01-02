# zen-sdk/pkg/metrics

Prometheus metrics helpers for zen components.

## Overview

This package provides standardized metrics recording for:
- **Kubernetes Controllers**: Reconciliation metrics via `Recorder`
- **HTTP Services**: HTTP request metrics via `HTTPMetrics`

## Features

- ✅ Controller reconciliation metrics (counter, duration histogram, errors)
- ✅ HTTP request metrics (counter, latency histogram, inflight gauge)
- ✅ Supports both controller-runtime metrics registry and prometheus default registry
- ✅ Configurable metric names and labels
- ✅ Thread-safe metric registration

## Usage

### Controller Reconciliation Metrics

For Kubernetes controllers (zen-lock, zen-flow, zen-gc, zen-lead, zen-watcher, zen-ingester, zen-egress):

```go
import "github.com/kube-zen/zen-sdk/pkg/metrics"

// Create recorder (automatically registers with controller-runtime metrics.Registry)
recorder := metrics.NewRecorder("zen-gc")

// Record successful reconciliation
duration := time.Since(start).Seconds()
recorder.RecordReconciliationSuccess(duration)

// Record failed reconciliation
recorder.RecordReconciliationError(duration)

// Record custom error
recorder.RecordError("webhook")
```

**Metrics Exposed:**
- `zen_reconciliations_total{component="zen-gc", result="success|error"}` - Counter
- `zen_reconciliation_duration_seconds{component="zen-gc", result="success|error"}` - Histogram
- `zen_errors_total{component="zen-gc", type="reconciliation|webhook|..."}` - Counter

### HTTP Request Metrics

For HTTP services (zen-back, zen-bff, zen-websocket, zen-bridge, etc.):

```go
import (
    "github.com/kube-zen/zen-sdk/pkg/metrics"
    "net/http"
)

// Create HTTP metrics (uses controller-runtime registry by default)
httpMetrics, err := metrics.NewHTTPMetrics(metrics.HTTPMetricsConfig{
    Component: "zen-back",
    Prefix:    "zen_http", // optional, defaults to "zen_http"
})
if err != nil {
    log.Fatal(err)
}

// Use as middleware
router.Use(httpMetrics.Middleware("api-v1"))

// Or wrap specific handlers
handler := httpMetrics.Middleware("health")(healthHandler)
```

**Metrics Exposed:**
- `zen_http_requests_total{component="zen-back", route="api-v1", method="GET", code="200"}` - Counter
- `zen_http_request_duration_ms{component="zen-back", route="api-v1", method="GET", code="200"}` - Histogram
- `zen_http_requests_inflight{component="zen-back", route="api-v1"}` - Gauge

### Custom Registry

To use a custom Prometheus registry (instead of controller-runtime):

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/kube-zen/zen-sdk/pkg/metrics"
)

// Create custom registry
registry := prometheus.NewRegistry()

// Use with HTTP metrics
httpMetrics, err := metrics.NewHTTPMetrics(metrics.HTTPMetricsConfig{
    Component: "zen-back",
    Registry:  registry, // Use custom registry
})
```

## Registry Patterns

This package supports multiple registry patterns:

1. **Controller-runtime Registry** (default for controllers)
   - Used by: zen-lock, zen-flow, zen-gc, zen-lead, zen-watcher, zen-ingester, zen-egress
   - Accessible at `/metrics` endpoint via controller-runtime metrics server

2. **Prometheus Default Registry** (for HTTP services)
   - Use custom registry if you need separate metrics
   - Expose via `promhttp.Handler()` on `/metrics` endpoint

3. **Custom Registry**
   - Pass `Registry` in config to use custom registry
   - Useful for testing or component-specific metrics isolation

## Metric Naming Conventions

- **Prefix**: Component-specific prefix (e.g., `zen_http`, `zen_reconciliations`)
- **Component Label**: Use `ConstLabels` (not variable labels) to reduce cardinality
- **Standard Labels**: `route`, `method`, `code` for HTTP; `result`, `type` for controllers
- **Units**: Duration in milliseconds (`_ms`) or seconds (`_seconds`), size in bytes (`_bytes`)

## Best Practices

1. **Use ConstLabels for Component**: Reduces cardinality, improves performance
2. **Keep Labels Minimal**: Only add labels you'll query/filter on
3. **Standard Buckets**: Use default buckets unless you have specific requirements
4. **Registry Choice**: Use controller-runtime registry for controllers, custom registry for HTTP services if needed

## Integration Examples

### zen-back (SaaS HTTP Service)

```go
// Current: Custom metrics in middleware/observability.go
// Future: Migrate to zen-sdk/pkg/metrics HTTPMetrics
httpMetrics, _ := metrics.NewHTTPMetrics(metrics.HTTPMetricsConfig{
    Component: "zen-back",
    Registry:  prometheus.DefaultRegisterer, // or custom registry
})
```

### zen-bridge (SaaS HTTP Service)

```go
// Current: promauto in registry_metrics.go
// Future: Consider HTTPMetrics for HTTP endpoints
```

### zen-lead (OSS Controller)

```go
// Current: Custom metrics with promauto
// Note: zen-lead uses zen-sdk/pkg/metrics.Recorder for reconciliation
//       but has additional custom metrics in pkg/metrics/metrics.go
```

## Migration Guide

To migrate from custom metrics to zen-sdk:

1. **Identify metrics type**: HTTP vs Controller reconciliation
2. **Create appropriate recorder**: `NewRecorder()` or `NewHTTPMetrics()`
3. **Replace metric recording**: Use SDK methods instead of direct prometheus calls
4. **Update registry**: Use SDK's registry handling
5. **Test**: Verify metrics are exposed correctly at `/metrics`

## Metrics Conventions

See [CONVENTIONS.md](CONVENTIONS.md) for detailed metrics conventions including:
- Naming conventions (prefixes, units, suffixes)
- Label conventions (standard labels, cardinality)
- Metric types (Counter, Histogram, Gauge)
- Registry patterns
- Common patterns and anti-patterns

## See Also

- [Metrics Conventions](CONVENTIONS.md) - Detailed conventions and best practices
- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
- [controller-runtime Metrics](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/metrics)

