# HTTP Client Package

**Hardened HTTP client with connection pooling, rate limiting, and structured logging**

## Overview

This package provides a hardened HTTP client implementation that can be used across OSS components (zen-watcher, zen-flow, etc.) for consistent HTTP request handling with proper connection pooling, rate limiting, and structured logging.

## Features

- ✅ **Connection Pooling**: Configurable connection pool settings (MaxIdleConns, MaxConnsPerHost, IdleConnTimeout)
- ✅ **Rate Limiting**: Optional rate limiting using `golang.org/x/time/rate`
- ✅ **Structured Logging**: Integration with `zen-sdk/pkg/logging` for request/response logging
- ✅ **TLS Configuration**: Support for custom TLS config or insecure skip verify (development only)
- ✅ **Environment Configuration**: Configurable via environment variables with sensible defaults
- ✅ **Timeout Management**: Configurable timeouts for requests, TLS handshake, response headers

## Quick Start

```go
import (
    "github.com/kube-zen/zen-sdk/pkg/http"
    "context"
)

// Create client with defaults (uses environment variables)
client := http.NewClient(nil)

// Or with custom configuration
config := &http.ClientConfig{
    Timeout: 30 * time.Second,
    MaxIdleConns: 100,
    RateLimitEnabled: true,
    RateLimitRPS: 10.0,
    ServiceName: "my-service",
}
client := http.NewClient(config)

// Make requests
resp, err := client.Get(ctx, "https://api.example.com/data")
if err != nil {
    return err
}
defer resp.Body.Close()

// Or POST
resp, err := client.Post(ctx, "https://api.example.com/data", "application/json", body)
```

## Configuration

### Environment Variables

- `HTTP_TIMEOUT` - Request timeout (default: `30s`)
- `HTTP_MAX_IDLE_CONNS` - Maximum idle connections (default: `100`)
- `HTTP_MAX_CONNS_PER_HOST` - Maximum connections per host (default: `10`)
- `HTTP_IDLE_CONN_TIMEOUT` - Idle connection timeout (default: `90s`)

### Programmatic Configuration

```go
config := &http.ClientConfig{
    Timeout:               30 * time.Second,
    MaxIdleConns:          100,
    MaxConnsPerHost:       10,
    IdleConnTimeout:       90 * time.Second,
    TLSInsecureSkipVerify: false, // Only for development/testing
    RateLimitEnabled:      true,
    RateLimitRPS:          10.0,
    RateLimitBurst:        10,
    LoggingEnabled:        true,
    ServiceName:           "my-service",
}
```

## Usage Examples

### Basic GET Request

```go
client := http.NewClient(nil)
resp, err := client.Get(ctx, "https://api.example.com/data")
if err != nil {
    return err
}
defer resp.Body.Close()
```

### POST Request with Body

```go
body := strings.NewReader(`{"key": "value"}`)
resp, err := client.Post(ctx, "https://api.example.com/data", "application/json", body)
```

### Custom Request

```go
req, err := http.NewRequestWithContext(ctx, "PUT", "https://api.example.com/data", body)
if err != nil {
    return err
}
req.Header.Set("Authorization", "Bearer token")
resp, err := client.Do(req)
```

### Rate Limited Client

```go
config := &http.ClientConfig{
    RateLimitEnabled: true,
    RateLimitRPS:     5.0,  // 5 requests per second
    RateLimitBurst:  10,    // Burst of 10 requests
}
client := http.NewClient(config)
```

## Migration from zen-watcher

If you're migrating from `zen-watcher/pkg/http`:

**Before:**
```go
import "github.com/kube-zen/zen-watcher/pkg/http"

client := http.NewHardenedHTTPClient(nil)
resp, err := client.Get(ctx, url)
```

**After:**
```go
import sdkhttp "github.com/kube-zen/zen-sdk/pkg/http"

client := sdkhttp.NewClient(nil)
resp, err := client.Get(ctx, url)
```

## Related

- [zen-sdk/pkg/config](../config/README.md) - Configuration helpers
- [zen-sdk/pkg/logging](../logging/README.md) - Structured logging
- [zen-sdk/pkg/retry](../retry/README.md) - Retry logic (can be combined with HTTP client)

