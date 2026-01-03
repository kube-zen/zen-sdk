# Kubernetes Resource Name Generation Package

Common Kubernetes resource name generation and truncation utilities extracted from OSS components.

## Features

- **Name Generation**: Generate names with prefixes and suffixes
- **Name Truncation**: Truncate names to Kubernetes limits while preserving critical suffixes
- **Hash-based Names**: Generate deterministic names with hash suffixes for uniqueness
- **DNS-1123 Compliance**: Ensures generated names comply with DNS-1123 requirements

## Usage

### Generate Name with Suffix

```go
import "github.com/kube-zen/zen-sdk/pkg/k8s/names"

// Generate a name with prefix and suffix, truncating if needed
name := names.GenerateName("zen-lock-inject-namespace-pod", "hash123", 253)
```

### Truncate Name

```go
// Truncate a name to max length, ensuring it ends with alphanumeric
truncated := names.TruncateName("very-long-resource-name", 253)
```

### Truncate Name Preserving Suffix

```go
// Truncate while preserving a critical suffix (e.g., hash)
name := names.TruncateNamePreserveSuffix("prefix-hash123", "hash123", 253)
```

### Generate Name with Hash

```go
// Generate a deterministic name with hash suffix
name := names.GenerateNameWithHash("zen-lock-inject", "namespace-pod", 253)
```

## Constants

- `MaxResourceNameLength` (253): Maximum length for Kubernetes resource names
- `MaxLabelValueLength` (63): Maximum length for label values

## Migration

### From zen-lock

```go
// Before
func GenerateSecretName(namespace, podName string) string {
    secretNameBase := fmt.Sprintf("zen-lock-inject-%s-%s", namespace, podName)
    hash := sha256.Sum256([]byte(secretNameBase))
    hashStr := hex.EncodeToString(hash[:])[:16]
    prefix := fmt.Sprintf("zen-lock-inject-%s-%s-", namespace, podName)
    secretName := prefix + hashStr
    // ... truncation logic ...
    return secretName
}

// After
import (
    sdknames "github.com/kube-zen/zen-sdk/pkg/k8s/names"
)

func GenerateSecretName(namespace, podName string) string {
    base := fmt.Sprintf("zen-lock-inject-%s-%s", namespace, podName)
    prefix := fmt.Sprintf("zen-lock-inject-%s-%s-", namespace, podName)
    return sdknames.GenerateNameWithHash(prefix, base, sdknames.MaxResourceNameLength)
}
```

## Extraction Status

- ✅ **Name Generation**: Extracted from zen-lock
- ✅ **Hash-based Names**: Extracted from zen-lock
- ✅ **Truncation with Suffix Preservation**: Extracted from zen-lock

