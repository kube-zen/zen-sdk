# Zen SDK

**Shared library for cross-cutting concerns across Zen tools.**

Zen SDK provides reusable components for Kubernetes operators and controllers, eliminating code duplication across zen-flow, zen-lock, zen-watcher, and other Zen tools.

## Philosophy

**Do not create a monorepo. Create a shared library.**

- ✅ **Modular**: Each tool is a separate repository
- ✅ **Lightweight**: Import only what you need
- ✅ **DRY**: Write once, use everywhere
- ✅ **Versioned**: Independent versioning per tool

## Components

### `pkg/leader` - Leader Election

Wrapper around controller-runtime's built-in leader election. Provides a simple, consistent API for enabling HA.

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

```go
import "github.com/kube-zen/zen-sdk/pkg/metrics"

recorder := metrics.NewRecorder("zen-flow")
recorder.RecordReconciliation("success")
```

### `pkg/logging` - Structured Logging

Consistent structured logging configuration across all tools.

```go
import "github.com/kube-zen/zen-sdk/pkg/logging"

logger := logging.NewLogger("zen-flow")
logger.Info("Controller started")
```

### `pkg/webhook` - Webhook Helpers

TLS certificate helpers and Kubernetes patch generation utilities.

```go
import "github.com/kube-zen/zen-sdk/pkg/webhook"

patch := webhook.GeneratePatch(obj, updates)
```

## Usage

### In zen-flow

```go
import (
    "github.com/kube-zen/zen-sdk/pkg/leader"
    "github.com/kube-zen/zen-sdk/pkg/metrics"
    "github.com/kube-zen/zen-sdk/pkg/logging"
)

func main() {
    logger := logging.NewLogger("zen-flow")
    
    opts := leader.Options{
        LeaseName: "zen-flow-controller",
        Enable: true,
    }
    
    mgr, err := ctrl.NewManager(cfg, leader.Setup(opts))
    // ...
}
```

### In zen-lock

```go
import (
    "github.com/kube-zen/zen-sdk/pkg/leader"
    "github.com/kube-zen/zen-sdk/pkg/webhook"
)

func main() {
    opts := leader.Options{
        LeaseName: "zen-lock-webhook",
        Enable: true,
    }
    
    mgr, err := ctrl.NewManager(cfg, leader.Setup(opts))
    // ...
}
```

## Installation

```bash
go get github.com/kube-zen/zen-sdk@latest
```

## Versioning

Zen SDK follows semantic versioning. Each Zen tool can depend on different versions:

- `zen-flow` might use `zen-sdk v1.0.0`
- `zen-lock` might use `zen-sdk v1.1.0`
- `zen-watcher` might use `zen-sdk v1.0.0`

This allows independent evolution while sharing common code.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache License 2.0 - See [LICENSE](LICENSE) file.

---

**Remember**: This is a library, not a monorepo. Each tool remains independent.

