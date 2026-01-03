# Kubernetes Metadata Utilities Package

Common Kubernetes metadata manipulation utilities extracted from OSS components.

## Features

- **GitOps Label/Annotation Filtering**: Remove GitOps tracking labels and annotations when copying metadata to generated resources
- **Owner Reference Management**: Set, check, and manage owner references on dependent resources
- **Label/Annotation Copying**: Copy labels and annotations between resources with optional GitOps filtering

## Usage

### Filter GitOps Labels

```go
import "github.com/kube-zen/zen-sdk/pkg/k8s/metadata"

// Filter GitOps tracking labels when copying metadata
filteredLabels := metadata.FilterGitOpsLabels(sourceLabels)
```

### Filter GitOps Annotations

```go
// Filter GitOps tracking annotations when copying metadata
filteredAnnotations := metadata.FilterGitOpsAnnotations(sourceAnnotations)
```

### Set Owner Reference

```go
import (
    sdkmetadata "github.com/kube-zen/zen-sdk/pkg/k8s/metadata"
    "k8s.io/apimachinery/pkg/runtime"
)

// Set owner reference on a dependent resource
err := sdkmetadata.SetOwnerReference(owner, dependent, scheme)
```

### Check Owner Reference

```go
// Check if an object has an owner reference
hasRef := sdkmetadata.HasOwnerReference(dependent, ownerUID)

// Get owner reference
ref := sdkmetadata.GetOwnerReference(dependent, ownerUID)

// Remove owner reference
removed := sdkmetadata.RemoveOwnerReference(dependent, ownerUID)
```

### Ensure Owner Reference

```go
// Ensure owner reference exists (adds if missing, updates if different)
added, err := sdkmetadata.EnsureOwnerReference(owner, dependent, scheme)
```

### Copy Labels and Annotations

```go
// Copy labels with GitOps filtering
sdkmetadata.CopyLabels(source, target, true)

// Copy annotations with GitOps filtering
sdkmetadata.CopyAnnotations(source, target, true)

// Copy without filtering
sdkmetadata.CopyLabels(source, target, false)
```

## GitOps Tracking Labels/Annotations

The package filters the following GitOps tracking labels and annotations:

**Labels:**
- `app.kubernetes.io/instance`
- `app.kubernetes.io/managed-by`
- `app.kubernetes.io/part-of`
- `app.kubernetes.io/version`
- `argocd.argoproj.io/instance`
- `fluxcd.io/part-of`
- `kustomize.toolkit.fluxcd.io/name`
- `kustomize.toolkit.fluxcd.io/namespace`
- `kustomize.toolkit.fluxcd.io/revision`

**Annotations:**
- `argocd.argoproj.io/sync-wave`
- `argocd.argoproj.io/sync-options`
- `fluxcd.io/sync-checksum`
- `kustomize.toolkit.fluxcd.io/checksum`

## Migration

### From zen-lead

```go
// Before
func filterGitOpsLabels(labels map[string]string) map[string]string {
    // ... local implementation ...
}

// After
import sdkmetadata "github.com/kube-zen/zen-sdk/pkg/k8s/metadata"

filteredLabels := sdkmetadata.FilterGitOpsLabels(labels)
```

## Extraction Status

- ✅ **GitOps Filtering**: Extracted from zen-lead
- ✅ **Owner Reference Management**: Generic helpers for all controllers
- ✅ **Label/Annotation Copying**: Generic helpers for metadata copying

