# Zen SDK Architecture

## Overview

Zen SDK is a **shared library** for cross-cutting concerns across Zen tools. It provides reusable components without creating a monorepo.

## Design Principles

1. **Modularity**: Each tool is a separate repository
2. **Lightweight**: Import only what you need
3. **DRY**: Write once, use everywhere
4. **Versioned**: Independent versioning per tool

## Package Structure

### `pkg/leader`

Wrapper around controller-runtime's built-in leader election.

**Purpose**: Standardize leader election configuration across all tools.

**Usage**:
```go
import "github.com/kube-zen/zen-sdk/pkg/leader"

opts := leader.Options{
    LeaseName: "zen-flow-controller",
    Enable: true,
}
mgr, err := ctrl.NewManager(cfg, leader.Setup(opts))
```

**Benefits**:
- Consistent configuration
- Simple API
- Uses controller-runtime's proven leader election

### `pkg/metrics`

Prometheus metrics recorder for Kubernetes controllers.

**Purpose**: Standard metrics across all tools.

**Usage**:
```go
import "github.com/kube-zen/zen-sdk/pkg/metrics"

recorder := metrics.NewRecorder("zen-flow")
recorder.RecordReconciliationSuccess(0.5)
```

**Metrics Provided**:
- `zen_reconciliations_total` - Total reconciliations
- `zen_reconciliation_duration_seconds` - Reconciliation duration
- `zen_errors_total` - Error counts

### `pkg/logging`

Structured logging configuration.

**Purpose**: Consistent logging across all tools.

**Usage**:
```go
import "github.com/kube-zen/zen-sdk/pkg/logging"

logger := logging.NewLogger("zen-flow")
logger.Info("Controller started")
```

**Features**:
- Structured logging with zap
- Component name context
- Development mode detection

### `pkg/webhook`

Webhook helpers for TLS and patch generation.

**Purpose**: Common webhook utilities.

**Usage**:
```go
import "github.com/kube-zen/zen-sdk/pkg/webhook"

patch, err := webhook.GeneratePatch(obj, updates)
```

**Features**:
- JSON patch generation
- TLS secret validation
- NamespacedName extraction

## Dependency Graph

```
zen-flow  ──┐
zen-lock  ──┼──> zen-sdk (shared library)
zen-watcher─┘
```

Each tool imports zen-sdk as a Go module dependency. They remain independent repositories.

## Versioning Strategy

Zen SDK follows semantic versioning. Tools can depend on different versions:

- `zen-flow` → `zen-sdk v1.0.0`
- `zen-lock` → `zen-sdk v1.1.0`
- `zen-watcher` → `zen-sdk v1.0.0`

This allows:
- Independent tool releases
- Gradual SDK adoption
- Backward compatibility

## What NOT to Include

❌ **Business Logic**: Belongs in individual tools
❌ **Tool-Specific Code**: Belongs in that tool's repo
❌ **CRDs**: Belongs in individual tools
❌ **Controllers**: Belongs in individual tools

## Migration Path

### For Existing Tools

1. **Add zen-sdk dependency:**
   ```bash
   go get github.com/kube-zen/zen-sdk@latest
   ```

2. **Replace custom code:**
   ```go
   // Before
   // Custom leader election code (50 lines)
   
   // After
   import "github.com/kube-zen/zen-sdk/pkg/leader"
   // Use leader.Setup(opts)
   ```

3. **Remove duplicate code:**
   - Delete custom leader election
   - Delete custom metrics setup
   - Delete custom logging

### For New Tools

1. Start with zen-sdk from the beginning
2. Import what you need
3. Focus on business logic, not infrastructure

## Benefits

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

## Future Enhancements

- `pkg/health` - Health check helpers
- `pkg/tracing` - OpenTelemetry integration
- `pkg/config` - Configuration management
- `pkg/client` - Kubernetes client helpers

---

**Remember**: Keep it focused on cross-cutting concerns. Business logic stays in tools.

