# Metrics Conventions

Standard conventions for Prometheus metrics across all zen components.

## Naming Conventions

### Metric Names

- **Format**: `{prefix}_{metric_name}_{unit?}_{type}`
- **Prefix**: Component-specific prefix (e.g., `zen_http`, `zen_reconciliations`, `zen_bridge`)
- **Metric Name**: Descriptive, lowercase, underscore-separated
- **Unit Suffix**: Optional unit suffix (`_ms`, `_seconds`, `_bytes`, `_total`)
- **Type Suffix**: Optional type indicator (`_total` for counters, `_rate` for rates)

**Examples:**
- `zen_http_requests_total` ✅
- `zen_reconciliation_duration_seconds` ✅
- `zen_bridge_registry_operations_total` ✅
- `httpRequestCounter` ❌ (camelCase, no prefix)

### Prefix Guidelines

- **Controllers**: Use component name prefix (e.g., `zen_gc`, `zen_flow`)
- **HTTP Services**: Use `zen_http` for standard HTTP metrics
- **Component-Specific**: Use component prefix for business metrics (e.g., `zen_bridge_registry_operations_total`)

## Label Conventions

### Standard Labels

#### HTTP Metrics
- `route` - API route/endpoint name (e.g., `api-v1`, `health`)
- `method` - HTTP method (e.g., `GET`, `POST`, `PUT`)
- `code` - HTTP status code (e.g., `200`, `404`, `500`)
- `component` - Component name (use ConstLabel, not variable label)

#### Controller Metrics
- `result` - Operation result (e.g., `success`, `error`)
- `type` - Error type (e.g., `reconciliation`, `webhook`)
- `component` - Component name (use ConstLabel)

#### Optional Labels
- `tenant_id` - For multi-tenant metrics (be careful with cardinality!)
- `cluster_id` - For cluster-scoped metrics
- `status` - Operation status (e.g., `created`, `duplicate`, `error`)

### Label Best Practices

1. **Use ConstLabels for Component**: Reduces cardinality, improves performance
   ```go
   ConstLabels: prometheus.Labels{"component": "zen-back"}
   ```

2. **Minimize Cardinality**: Avoid high-cardinality labels (user IDs, request IDs)
   - ✅ Good: `tenant_id` (limited number of tenants)
   - ❌ Bad: `request_id` (unlimited cardinality)

3. **Keep Labels Consistent**: Use same label names across related metrics
   - HTTP metrics: `route`, `method`, `code`
   - Controller metrics: `result`, `type`

4. **Document Label Values**: Document possible label values in metric help text
   ```go
   Help: "Total reconciliations (result: success|error)",
   ```

## Metric Types

### Counter (`CounterVec`)

Use for:
- Total counts (requests, operations, events)
- **Must** be monotonically increasing
- **Always** use `_total` suffix

```go
prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "zen_http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"route", "method", "code"},
)
```

### Histogram (`HistogramVec`)

Use for:
- Latency/duration measurements
- Request/response sizes
- **Always** specify units in name (`_ms`, `_seconds`, `_bytes`)

**Standard Buckets:**
- **HTTP Latency (ms)**: `[25, 50, 100, 200, 400, 800, 1600, 3200]`
- **Duration (seconds)**: `prometheus.ExponentialBuckets(0.001, 2, 10)` (1ms to ~1s)
- **Custom**: Define buckets based on expected range

```go
prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "zen_http_request_duration_ms",
        Help:    "HTTP request duration in milliseconds",
        Buckets: []float64{25, 50, 100, 200, 400, 800, 1600, 3200},
    },
    []string{"route", "method", "code"},
)
```

### Gauge (`GaugeVec`)

Use for:
- Current state (inflight requests, queue depth, connection pool size)
- Values that can increase or decrease

```go
prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "zen_http_requests_inflight",
        Help: "Current number of HTTP requests being processed",
    },
    []string{"route"},
)
```

## Units

Always specify units in metric names:

- **Time**: `_ms` (milliseconds), `_seconds` (seconds)
- **Size**: `_bytes`, `_kb`, `_mb`
- **Count**: `_total` (for counters), no suffix for gauges
- **Rate**: `_rate` (for rate metrics)

**Examples:**
- `zen_http_request_duration_ms` ✅
- `zen_reconciliation_duration_seconds` ✅
- `zen_cache_size_bytes` ✅
- `zen_http_request_duration` ❌ (no unit)

## Registry Patterns

### Controller Components

Use `controller-runtime/pkg/metrics.Registry`:
- Standard for all Kubernetes controllers
- Accessible via controller-runtime metrics server
- Use `zen-sdk/pkg/metrics.NewRecorder()` or `NewHTTPMetrics()` with `Registry: nil` (default)

```go
recorder := metrics.NewRecorder("zen-gc")
// Automatically uses controller-runtime metrics.Registry
```

### HTTP Services

Options:
1. **Controller-runtime Registry** (recommended for consistency)
   ```go
   httpMetrics, _ := metrics.NewHTTPMetrics(metrics.HTTPMetricsConfig{
       Component: "zen-back",
       Registry:  nil, // Uses controller-runtime metrics.Registry
   })
   ```

2. **Prometheus Default Registry** (if separate metrics server)
   ```go
   httpMetrics, _ := metrics.NewHTTPMetrics(metrics.HTTPMetricsConfig{
       Component: "zen-back",
       Registry:  prometheus.DefaultRegisterer,
   })
   ```

3. **Custom Registry** (for testing or isolation)
   ```go
   registry := prometheus.NewRegistry()
   httpMetrics, _ := metrics.NewHTTPMetrics(metrics.HTTPMetricsConfig{
       Component: "zen-back",
       Registry:  registry,
   })
   ```

## Common Patterns

### HTTP Request Metrics

```go
// Standard HTTP metrics (requests, latency, inflight)
httpMetrics, _ := metrics.NewHTTPMetrics(metrics.HTTPMetricsConfig{
    Component: "zen-back",
})
router.Use(httpMetrics.Middleware("api-v1"))
```

**Metrics:**
- `zen_http_requests_total{component="zen-back", route="api-v1", method="GET", code="200"}`
- `zen_http_request_duration_ms{component="zen-back", route="api-v1", method="GET", code="200"}`
- `zen_http_requests_inflight{component="zen-back", route="api-v1"}`

### Database Query Metrics

```go
// Custom DB metrics (component-specific)
dbQueriesTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "zen_back_db_queries_total",
        Help: "Total database queries by operation and table",
        ConstLabels: prometheus.Labels{"component": "zen-back"},
    },
    []string{"operation", "table", "status"},
)

dbQueryDuration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "zen_back_db_query_duration_ms",
        Help:    "Database query duration in milliseconds",
        Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
        ConstLabels: prometheus.Labels{"component": "zen-back"},
    },
    []string{"operation", "table"},
)
```

### Cache Metrics

```go
cacheOperationsTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "zen_back_cache_operations_total",
        Help: "Total cache operations by type and result",
        ConstLabels: prometheus.Labels{"component": "zen-back"},
    },
    []string{"operation", "result"}, // operation: get, set, delete; result: hit, miss, error
)
```

### Business Metrics

```go
// Tenant activity (be careful with cardinality!)
tenantActivityTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "zen_back_tenant_activity_total",
        Help: "Tenant activity by action type",
        ConstLabels: prometheus.Labels{"component": "zen-back"},
    },
    []string{"action", "tenant_id"}, // Only if tenant count is limited!
)
```

## Anti-Patterns

### ❌ High Cardinality Labels

```go
// BAD: request_id has unlimited cardinality
prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "zen_http_requests_total"},
    []string{"request_id"}, // ❌ Too many unique values!
)
```

### ❌ Inconsistent Naming

```go
// BAD: Inconsistent naming across components
prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "httpRequests"}, // ❌ No prefix, camelCase
    []string{"route"},
)
```

### ❌ Missing Units

```go
// BAD: No unit specified
prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "zen_http_request_duration", // ❌ Is this ms? seconds?
    },
    []string{"route"},
)
```

### ❌ Wrong Metric Type

```go
// BAD: Counter for a value that can decrease
prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "zen_http_inflight_requests_total"}, // ❌ Should be Gauge!
    []string{"route"},
)
```

## Migration Checklist

When migrating custom metrics to zen-sdk:

- [ ] Use `zen-sdk/pkg/metrics` helpers when possible
- [ ] Follow naming conventions (prefix, units, suffixes)
- [ ] Use ConstLabels for component name
- [ ] Minimize label cardinality
- [ ] Document metric purpose and label values
- [ ] Choose appropriate buckets for histograms
- [ ] Use correct metric type (Counter vs Gauge vs Histogram)
- [ ] Register with appropriate registry
- [ ] Test metrics are exposed correctly at `/metrics`

## References

- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
- [Prometheus Metric Types](https://prometheus.io/docs/concepts/metric_types/)
- [Prometheus Histograms and Summaries](https://prometheus.io/docs/practices/histograms/)

