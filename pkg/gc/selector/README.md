# Selector Package

Resource selector matching primitives for Kubernetes resources.

## Overview

This package provides APIs for matching Kubernetes resources based on labels, annotations, fields, and phases. Extracted from zen-gc for reuse across components.

## Usage

### Match Labels

```go
import "github.com/kube-zen/zen-sdk/pkg/gc/selector"

conditions := []selector.LabelCondition{
    {Key: "app", Value: "myapp", Operator: "Equals"},
    {Key: "env", Value: "prod", Operator: "Exists"},
}

if selector.MatchesLabels(resource, conditions) {
    fmt.Println("Resource matches label conditions")
}
```

### Match Annotations

```go
conditions := []selector.AnnotationCondition{
    {Key: "cleanup", Value: "enabled"},
}

if selector.MatchesAnnotations(resource, conditions) {
    fmt.Println("Resource has cleanup annotation")
}
```

### Match Phase

```go
// Match resources in Failed or Succeeded phase
phases := []string{"Failed", "Succeeded"}

if selector.MatchesPhase(resource, phases) {
    fmt.Println("Resource is in terminal phase")
}
```

### Match Field Conditions

```go
conditions := []selector.FieldCondition{
    {Path: "status.phase", Operator: "In", Values: []string{"Failed", "Succeeded"}},
    {Path: "spec.severity", Operator: "Equals", Value: "critical"},
}

if selector.MatchesFields(resource, conditions) {
    fmt.Println("Resource matches field conditions")
}
```

### Match All Conditions

```go
conditions := &selector.Conditions{
    Phase: []string{"Failed", "Succeeded"},
    HasLabels: []selector.LabelCondition{
        {Key: "app", Value: "myapp"},
    },
    And: []selector.FieldCondition{
        {Path: "spec.severity", Operator: "Equals", Value: "critical"},
    },
}

if selector.MatchesConditions(resource, conditions) {
    fmt.Println("Resource matches all conditions")
}
```

### Match Kubernetes Label Selector

```go
// Standard Kubernetes label selector syntax
selectorStr := "app=myapp,env in (prod,staging)"

matches, err := selector.MatchesLabelSelector(resource, selectorStr)
if err == nil && matches {
    fmt.Println("Resource matches label selector")
}
```

## API Reference

### Types

#### `LabelCondition`

```go
type LabelCondition struct {
    Key      string
    Value    string
    Operator string // Exists, Equals, In, NotIn
}
```

#### `AnnotationCondition`

```go
type AnnotationCondition struct {
    Key   string
    Value string
}
```

#### `FieldCondition`

```go
type FieldCondition struct {
    Path     string   // Dot-separated field path
    Operator string   // Equals, In, NotIn
    Value    string   // Single value
    Values   []string // Multiple values (for In/NotIn)
}
```

#### `Conditions`

```go
type Conditions struct {
    Phase          []string              // OR logic
    HasLabels      []LabelCondition      // AND logic
    HasAnnotations []AnnotationCondition // AND logic
    And            []FieldCondition      // AND logic
}
```

### Functions

- `MatchesLabels(resource, conditions)` - Match label conditions (AND)
- `MatchesAnnotations(resource, conditions)` - Match annotation conditions (AND)
- `MatchesPhase(resource, phases)` - Match phase (OR)
- `MatchesField(resource, condition)` - Match single field condition
- `MatchesFields(resource, conditions)` - Match field conditions (AND)
- `MatchesConditions(resource, conditions)` - Match all conditions (AND)
- `MatchesLabelSelector(resource, selectorStr)` - Match Kubernetes label selector

## Operators

### Label Operators

- `Exists`: Label key must exist (value ignored)
- `Equals`: Label key must exist and value must match
- `In`: Label value must match the specified value
- `NotIn`: Label value must NOT match the specified value

### Field Operators

- `Equals`: Field value must equal the specified value
- `In`: Field value must be in the Values list
- `NotIn`: Field value must NOT be in the Values list

## Logic

- **Phase**: OR logic (matches if any phase matches)
- **Labels**: AND logic (all conditions must match)
- **Annotations**: AND logic (all conditions must match)
- **Fields**: AND logic (all conditions must match)
- **Overall**: AND logic (all condition types must match)

## Integration

This package is used by:
- `zen-gc` - Resource filtering before GC
- `zen-watcher` - Observation filtering
- Any component needing resource matching

## Testing

```bash
cd pkg/gc/selector
go test -v ./...
```

