# Zen SDK

**Shared library for cross-cutting concerns across Zen tools.**

Zen SDK provides reusable components for Kubernetes operators and controllers, eliminating code duplication across zen-flow, zen-lock, zen-watcher, and other Zen tools.

## Philosophy

**Do not create a monorepo. Create a shared library.**

- ✅ **Modular**: Each tool is a separate repository
- ✅ **Lightweight**: Import only what you need
- ✅ **DRY**: Write once, use everywhere
- ✅ **Versioned**: Independent versioning per tool

## Quick Start

```bash
go get github.com/kube-zen/zen-sdk@latest
```

```go
import (
    "github.com/kube-zen/zen-sdk/pkg/leader"
    "github.com/kube-zen/zen-sdk/pkg/metrics"
    "github.com/kube-zen/zen-sdk/pkg/logging"
)

// Enable leader election
opts := leader.Options{
    LeaseName: "my-controller",
    Enable:    true,
}
mgr, err := ctrl.NewManager(cfg, ctrl.Options{}, leader.Setup(opts))

// Record metrics
recorder := metrics.NewRecorder("my-controller")
recorder.RecordReconciliationSuccess(0.5)

// Use logging
logger := logging.NewLogger("my-controller")
logger.Info("Controller started")
```

## Components

### `pkg/leader` - Leader Election

Wrapper around controller-runtime's built-in leader election. Provides a simple, consistent API for enabling HA.

**Usage:**
```go
import "github.com/kube-zen/zen-sdk/pkg/leader"

opts := leader.Options{
    LeaseName: "zen-flow-controller",
    Enable: true,
}
manager := ctrl.NewManager(..., leader.Setup(opts))
```

### `pkg/metrics` - Prometheus Metrics

Standard Prometheus metrics setup and common metrics for Kubernetes controllers.

**Usage:**
```go
import "github.com/kube-zen/zen-sdk/pkg/metrics"

recorder := metrics.NewRecorder("zen-flow")
recorder.RecordReconciliation("success", 0.5)
```

**Metrics:**
- `zen_reconciliations_total{component, result}` - Total reconciliations
- `zen_reconciliation_duration_seconds{component, result}` - Duration histogram
- `zen_errors_total{component, type}` - Error counts

### `pkg/logging` - Structured Logging

Consistent structured logging configuration across all tools.

**Usage:**
```go
import "github.com/kube-zen/zen-sdk/pkg/logging"

logger := logging.NewLogger("zen-flow")
logger.Info("Controller started")
logger.WithField("namespace", "default").Info("Processing")
```

### `pkg/webhook` - Webhook Helpers

TLS certificate helpers and Kubernetes patch generation utilities.

**Usage:**
```go
import "github.com/kube-zen/zen-sdk/pkg/webhook"

patch := webhook.GeneratePatch(obj, updates)
// or
patch := webhook.GenerateAddPatch("/metadata/labels/test", "value")
```

### `pkg/lifecycle` - Graceful Shutdown

Standardized graceful shutdown helpers for HTTP servers, gRPC servers, and worker services.

**Usage:**
```go
import "github.com/kube-zen/zen-sdk/pkg/lifecycle"

// Create shutdown context
ctx, cancel := lifecycle.ShutdownContext(context.Background(), "my-component")
defer cancel()

// Start server
go server.ListenAndServe()

// Wait for shutdown signal
<-ctx.Done()

// Graceful shutdown
lifecycle.ShutdownHTTPServer(ctx, server, "my-component", 30*time.Second)
```

**Features:**
- Signal handling (`SIGINT`, `SIGTERM`)
- HTTP server graceful shutdown
- gRPC server graceful shutdown
- Worker service coordination
- Structured logging integration

### `pkg/gc` - Garbage Collection Primitives

Shared GC evaluation primitives extracted from zen-gc and zen-watcher.

**Packages:**
- `ratelimiter` - Token bucket rate limiting
- `backoff` - Exponential backoff for retries
- `ttl` - TTL (Time-To-Live) evaluation
- `fieldpath` - Field path parsing and value extraction
- `selector` - Resource selector matching (labels, annotations, fields)

**Usage:**
```go
import (
    "github.com/kube-zen/zen-sdk/pkg/gc/ttl"
    "github.com/kube-zen/zen-sdk/pkg/gc/ratelimiter"
)

// TTL evaluation
ttlSeconds := int64(3600)
spec := &ttl.Spec{SecondsAfterCreation: &ttlSeconds}
expired, _ := ttl.IsExpired(resource, spec)

// Rate limiting
rl := ratelimiter.NewRateLimiter(10) // 10 ops/sec
rl.Wait(ctx)
```

## Migration Guide

See [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) for detailed migration instructions.

**Quick migration for zen-flow:**

```go
// Before
func ManagerOptions(...) ctrl.Options {
    // 15 lines of custom leader election code
}

// After
import "github.com/kube-zen/zen-sdk/pkg/leader"
leaderOpts := leader.Options{LeaseName: "...", Enable: true}
mgr := ctrl.NewManager(..., leader.Setup(leaderOpts))
```

**Result:** 76% code reduction, consistent behavior, easier maintenance.

## Examples

- [Leader Election](examples/leader_example.go)
- [Metrics](examples/metrics_example.go)
- [Logging](examples/logging_example.go)
- [Webhook](examples/webhook_example.go)
- [Lifecycle/Graceful Shutdown](pkg/lifecycle/README.md)
- [zen-flow Migration](examples/zen-flow-migration.go)
- [zen-lock Migration](examples/zen-lock-migration.go)

## Documentation

- [Quick Start](QUICKSTART.md) - Get started in 5 minutes
- [API Reference](API_REFERENCE.md) - Complete API documentation
- [Architecture](ARCHITECTURE.md) - Design and architecture
- [Migration Guide](MIGRATION_GUIDE.md) - Migrate existing tools
- [Migration Examples](MIGRATION_EXAMPLES.md) - Practical examples
- [Contributing](CONTRIBUTING.md) - Contribution guidelines

## Versioning

Zen SDK follows semantic versioning. Each Zen tool can depend on different versions:

- `zen-flow` might use `zen-sdk v1.0.0`
- `zen-lock` might use `zen-sdk v1.1.0`
- `zen-watcher` might use `zen-sdk v1.0.0`

This allows independent evolution while sharing common code.

## Impact

### Before (Without SDK)

- zen-flow: 50 lines of leader election
- zen-lock: 50 lines of leader election
- zen-watcher: 50 lines of leader election
- **Total: 150 lines to maintain**

### After (With SDK)

- zen-sdk: 50 lines of leader election (written once)
- zen-flow: Import and use
- zen-lock: Import and use
- zen-watcher: Import and use
- **Total: 50 lines to maintain**

**Result: 3x code reduction, single source of truth**

## Installation

```bash
go get github.com/kube-zen/zen-sdk@latest
```

## Requirements

- Go 1.24+
- Kubernetes 1.26+
- controller-runtime v0.18.0+

## License

Apache License 2.0 - See [LICENSE](LICENSE) file.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

**Remember**: This is a library, not a monorepo. Each tool remains independent.
