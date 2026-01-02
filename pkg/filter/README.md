# Filter Package

Event filtering package for source-level observation filtering with expression and list-based rules.

## Features

- **Expression-Based Filtering**: SQL-like expressions (e.g., `severity >= HIGH AND category IN [security, compliance]`)
- **List-Based Filtering**: Include/exclude rules for severity, event types, namespaces, kinds, categories
- **Global Namespace Filtering**: Apply namespace filters across all sources
- **Per-Source Configuration**: Source-specific filter rules
- **Dynamic Configuration**: Thread-safe config updates without restart
- **Optional Metrics**: Interface for components to track filter decisions

## Usage

```go
import "github.com/kube-zen/zen-sdk/pkg/filter"

// Create filter with configuration
config := &filter.FilterConfig{
    Sources: map[string]filter.SourceFilter{
        "falco": {
            MinSeverity: "HIGH",
            ExcludeNamespaces: []string{"kube-system"},
        },
    },
}

f := filter.NewFilter(config)

// Check if observation should be allowed
observation := &unstructured.Unstructured{
    Object: map[string]interface{}{
        "spec": map[string]interface{}{
            "source":   "falco",
            "severity": "CRITICAL",
            // ... other fields
        },
    },
}

if f.Allow(observation) {
    // Process observation
}

// Update configuration dynamically
newConfig := &filter.FilterConfig{...}
f.UpdateConfig(newConfig)
```

## Expression Syntax

Expressions support:

- **Comparisons**: `=`, `!=`, `>`, `>=`, `<`, `<=`
- **String Operations**: `CONTAINS`, `STARTS_WITH`, `ENDS_WITH`
- **Set Operations**: `IN`, `NOT IN`
- **Existence**: `EXISTS`, `NOT EXISTS`
- **Logical**: `AND`, `OR`, `NOT`
- **Macros**: `is_critical`, `is_high`, `is_security`, `is_compliance`

Example: `(severity >= HIGH) AND (category IN [security, compliance])`

## Metrics Interface

Components can implement `FilterMetrics` to track filter decisions:

```go
type FilterMetrics interface {
    RecordFilterDecision(source, decision, reason string)
    RecordEvaluationDuration(source, ruleType string, durationSeconds float64)
}
```

## Thread Safety

All methods are thread-safe and can be called concurrently.

