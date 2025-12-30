# Zen SDK API Reference

Quick reference for all zen-sdk packages.

## pkg/leader

### Types

```go
type Options struct {
    LeaseName     string        // Name of the Lease resource
    Enable        bool          // Enable leader election
    Namespace     string        // Namespace for Lease (optional)
    LeaseDuration time.Duration // How long leader holds lease (default: 15s)
    RenewDeadline time.Duration // Time to renew before losing (default: 10s)
    RetryPeriod   time.Duration // How often to retry (default: 2s)
}
```

### Functions

```go
// DefaultOptions returns default options for a lease name
func DefaultOptions(leaseName string) Options

// Setup returns a function that configures leader election
func Setup(opts Options) func(*ctrl.Options)

// ManagerOptions returns ctrl.Options with leader election configured
func ManagerOptions(baseOpts ctrl.Options, leaderOpts Options) ctrl.Options
```

## pkg/metrics

### Types

```go
type Recorder struct {
    // Private fields
}
```

### Functions

```go
// NewRecorder creates a new metrics recorder
func NewRecorder(componentName string) *Recorder
```

### Methods

```go
// RecordReconciliation records a reconciliation attempt
func (r *Recorder) RecordReconciliation(result string, durationSeconds float64)

// RecordError records an error
func (r *Recorder) RecordError(errorType string)

// RecordReconciliationSuccess records a successful reconciliation
func (r *Recorder) RecordReconciliationSuccess(durationSeconds float64)

// RecordReconciliationError records a failed reconciliation
func (r *Recorder) RecordReconciliationError(durationSeconds float64)
```

### Metrics Exposed

- `zen_reconciliations_total{component, result}` - Total reconciliations
- `zen_reconciliation_duration_seconds{component, result}` - Duration histogram
- `zen_errors_total{component, type}` - Error counts

## pkg/logging

### Types

```go
type Logger struct {
    // Private fields
}
```

### Functions

```go
// NewLogger creates a new structured logger
func NewLogger(componentName string) *Logger
```

### Methods

```go
// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field)

// Error logs an error message
func (l *Logger) Error(err error, msg string, fields ...zap.Field)

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zap.Field)

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zap.Field)

// WithComponent adds component name to context
func (l *Logger) WithComponent(component string) *Logger

// WithField adds a field to context
func (l *Logger) WithField(key string, value interface{}) *Logger

// WithFields adds multiple fields to context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger
```

## pkg/webhook

### Types

```go
type Patch struct {
    Op    string      `json:"op"`    // "add", "remove", "replace"
    Path  string      `json:"path"`  // JSON path
    Value interface{} `json:"value,omitempty"`
}
```

### Functions

```go
// GeneratePatch generates a JSON patch from updates
func GeneratePatch(obj runtime.Object, updates map[string]interface{}) ([]byte, error)

// GenerateAddPatch generates a patch to add a field
func GenerateAddPatch(path string, value interface{}) ([]byte, error)

// GenerateRemovePatch generates a patch to remove a field
func GenerateRemovePatch(path string) ([]byte, error)

// ValidateTLSSecret validates TLS secret data
func ValidateTLSSecret(secret *unstructured.Unstructured) error

// GetNamespacedName extracts NamespacedName from an object
func GetNamespacedName(obj runtime.Object) (types.NamespacedName, error)
```

## Common Patterns

### Leader Election Pattern

```go
opts := leader.Options{
    LeaseName: "my-controller",
    Enable:    true,
}
mgr, err := ctrl.NewManager(cfg, ctrl.Options{}, leader.Setup(opts))
```

### Metrics Pattern

```go
recorder := metrics.NewRecorder("my-controller")
start := time.Now()
// ... work ...
recorder.RecordReconciliationSuccess(time.Since(start).Seconds())
```

### Logging Pattern

```go
logger := logging.NewLogger("my-controller")
logger.WithField("key", "value").Info("message")
```

### Webhook Pattern

```go
patch, err := webhook.GenerateAddPatch("/metadata/labels/test", "value")
return admission.Patched("", patch)
```

---

**See [examples/](examples/) for complete examples.**

