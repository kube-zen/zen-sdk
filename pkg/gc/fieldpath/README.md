# FieldPath Package

Field path evaluation primitives for Kubernetes resources.

## Overview

This package provides a simple API for extracting values from Kubernetes resources using dot-separated field paths (e.g., `spec.severity`, `status.lastProcessedAt`).

## Usage

### Get String Field

```go
import "github.com/kube-zen/zen-sdk/pkg/gc/fieldpath"

severity, found, err := fieldpath.GetString(resource, "spec.severity")
if found {
    fmt.Printf("Severity: %s\n", severity)
}
```

### Get Int64 Field

```go
ttl, found, err := fieldpath.GetInt64(resource, "spec.ttlSeconds")
if found {
    fmt.Printf("TTL: %d seconds\n", ttl)
}
```

### Get Bool Field

```go
enabled, found, err := fieldpath.GetBool(resource, "spec.enabled")
if found && enabled {
    fmt.Println("Feature is enabled")
}
```

### Check if Field Exists

```go
if fieldpath.Exists(resource, "status.conditions") {
    fmt.Println("Resource has conditions")
}
```

### Parse Field Path

```go
// Parse dot-separated path into slice
fields := fieldpath.Parse("spec.template.spec.containers")
// Returns: ["spec", "template", "spec", "containers"]
```

## API Reference

### Functions

- `GetString(resource, path)` - Get string value
- `GetInt64(resource, path)` - Get int64 value
- `GetBool(resource, path)` - Get boolean value
- `GetFloat64(resource, path)` - Get float64 value
- `Exists(resource, path)` - Check if field exists
- `Parse(path)` - Parse dot-separated path into slice

### Return Values

All `Get*` functions return `(value, found, error)`:
- `value`: The field value (zero value if not found)
- `found`: `true` if field exists, `false` otherwise
- `error`: Error during access (e.g., type mismatch)

## Examples

### Dynamic TTL Evaluation

```go
// Get TTL from resource field
ttl, found, _ := fieldpath.GetInt64(resource, "spec.ttlSeconds")
if found {
    expirationTime := resource.GetCreationTimestamp().Time.Add(time.Duration(ttl) * time.Second)
}
```

### Conditional Logic

```go
// Check severity level
severity, found, _ := fieldpath.GetString(resource, "spec.severity")
if found && severity == "critical" {
    // Handle critical resource
}
```

## Integration

This package is used by:
- `zen-sdk/pkg/gc/ttl` - TTL evaluation
- `zen-gc` - Garbage collection controller
- `zen-watcher` - Observation filtering

## Testing

```bash
cd pkg/gc/fieldpath
go test -v ./...
```

