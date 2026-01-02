# Kubernetes Validation Package

Common Kubernetes resource validation utilities extracted from OSS components.

## Features

- **DNS-1123 Subdomain Validation**: Validates resource names that can contain dots (e.g., API groups, resource names)
- **DNS-1123 Label Validation**: Validates resource names without dots (e.g., resource names, label keys/values)
- **GVR Validation**: Validates GroupVersionResource (API group, version, resource)
- **Resource Name Validation**: Validates resource names with optional length limits
- **Whitespace Checking**: Validates names without leading/trailing whitespace

## Usage

### DNS-1123 Subdomain Validation

```go
import "github.com/kube-zen/zen-sdk/pkg/k8s/validation"

// Validate a resource name that can contain dots
if err := validation.ValidateDNS1123Subdomain("my.resource.name"); err != nil {
    return err
}
```

### DNS-1123 Label Validation

```go
// Validate a resource name without dots
if err := validation.ValidateDNS1123Label("my-resource"); err != nil {
    return err
}
```

### Resource Name Validation with Length Limit

```go
// Validate with max length (e.g., 253 for annotations)
if err := validation.ValidateResourceName("my-resource", validation.MaxAnnotationValueLength); err != nil {
    return err
}

// Validate without length limit
if err := validation.ValidateResourceName("my-resource", 0); err != nil {
    return err
}
```

### GVR Validation

```go
import "k8s.io/apimachinery/pkg/runtime/schema"

gvr := schema.GroupVersionResource{
    Group:    "apps.kube-zen.io",
    Version:  "v1alpha1",
    Resource: "jobflows",
}

if err := validation.ValidateGVR(gvr); err != nil {
    return err
}

// Or validate from individual components
if err := validation.ValidateGVRConfig("apps.kube-zen.io", "v1alpha1", "jobflows"); err != nil {
    return err
}
```

### Name with Whitespace Check

```go
// Validates DNS-1123 subdomain and ensures no leading/trailing whitespace
if err := validation.ValidateNameWithWhitespaceCheck("my-resource"); err != nil {
    return err
}
```

## Constants

- `MaxAnnotationValueLength` (253): Maximum length for annotation values
- `MaxLabelValueLength` (63): Maximum length for label values
- `MaxResourceNameLength` (253): Maximum length for resource names

## Migration

### From zen-lock

```go
// Before
import "github.com/kube-zen/zen-lock/pkg/webhook"
if err := webhook.ValidateInjectAnnotation(name); err != nil {
    return err
}

// After
import "github.com/kube-zen/zen-sdk/pkg/k8s/validation"
if err := validation.ValidateResourceName(name, validation.MaxAnnotationValueLength); err != nil {
    return err
}
```

### From zen-watcher

```go
// Before
import "github.com/kube-zen/zen-watcher/pkg/config"
if err := config.ValidateGVR(gvr); err != nil {
    return err
}

// After
import "github.com/kube-zen/zen-sdk/pkg/k8s/validation"
if err := validation.ValidateGVR(gvr); err != nil {
    return err
}
```

### From zen-flow

```go
// Before
import "github.com/kube-zen/zen-flow/pkg/validation"
if err := validation.ValidateStepName(name); err != nil {
    return err
}

// After
import "github.com/kube-zen/zen-sdk/pkg/k8s/validation"
if err := validation.ValidateNameWithWhitespaceCheck(name); err != nil {
    return err
}
```

## Extraction Status

- ✅ **DNS-1123 Subdomain Validation**: Extracted from zen-lock, zen-flow
- ✅ **DNS-1123 Label Validation**: Extracted from zen-watcher
- ✅ **GVR Validation**: Extracted from zen-watcher
- ✅ **Resource Name Validation**: Extracted from zen-lock
- ✅ **Whitespace Checking**: Extracted from zen-flow

