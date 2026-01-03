# HTTP Client Package

**Hardened HTTP client with connection pooling, rate limiting, retry logic, Prometheus metrics, and structured logging**

## Overview

This package provides a hardened HTTP client implementation that can be used across OSS components (zen-watcher, zen-flow, etc.) for consistent HTTP request handling with proper connection pooling, rate limiting, retry logic, observability, and structured logging.

## Features

- ✅ **Connection Pooling**: Configurable connection pool settings (MaxIdleConns, MaxConnsPerHost, IdleConnTimeout)
- ✅ **Rate Limiting**: Optional rate limiting using `golang.org/x/time/rate`
- ✅ **Retry Logic**: Automatic retry with exponential backoff for network errors and retryable HTTP status codes
- ✅ **Prometheus Metrics**: Built-in metrics for requests, retries, errors, and latency
- ✅ **Structured Logging**: Integration with `zen-sdk/pkg/logging` for request/response logging
- ✅ **TLS Configuration**: Support for custom TLS config or insecure skip verify (development only)
- ✅ **Environment Configuration**: Configurable via environment variables with sensible defaults
- ✅ **Timeout Management**: Configurable timeouts for requests, TLS handshake, response headers
- ✅ **Request/Response Middleware**: Support for custom request and response processing
- ✅ **HTTP Methods**: Convenient methods for GET, POST, PUT, DELETE, and PostJSON

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
    RetryConfig:           retry.DefaultHTTPConfig(), // Retry configuration
    RateLimitEnabled:      true,
    RateLimitRPS:          10.0,
    RateLimitBurst:        10,
    LoggingEnabled:        true,
    ServiceName:           "my-service",
    // Optional middleware
    RequestMiddleware:     func(req *http.Request) error {
        req.Header.Set("X-Custom-Header", "value")
        return nil
    },
    ResponseMiddleware:    func(resp *http.Response) error {
        // Process response
        return nil
    },
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

### POST Request with JSON

```go
data := map[string]interface{}{
    "key": "value",
    "number": 42,
}
resp, err := client.PostJSON(ctx, "https://api.example.com/data", data)
```

### PUT Request

```go
body := strings.NewReader(`{"key": "updated"}`)
resp, err := client.Put(ctx, "https://api.example.com/data", "application/json", body)
```

### DELETE Request

```go
resp, err := client.Delete(ctx, "https://api.example.com/data/123")
```

### Custom Request

```go
req, err := http.NewRequestWithContext(ctx, "PATCH", "https://api.example.com/data", body)
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

### Client with Retry Configuration

```go
import "github.com/kube-zen/zen-sdk/pkg/retry"

retryCfg := retry.DefaultHTTPConfig()
retryCfg.MaxAttempts = 5
retryCfg.RetryOnStatusCodes = []int{429, 500, 502, 503, 504}

config := &http.ClientConfig{
    RetryConfig: retryCfg,
    ServiceName: "my-service",
}
client := http.NewClient(config)
```

### Prometheus Metrics

The HTTP client automatically exposes Prometheus metrics:

- `http_client_requests_total` - Total requests by service, method, status, and retry
- `http_client_request_duration_seconds` - Request duration histogram
- `http_client_retries_total` - Total retry attempts by service and reason
- `http_client_rate_limit_hits_total` - Rate limit hits by service
- `http_client_errors_total` - Errors by service and error type

These metrics are automatically registered with Prometheus when the package is imported.

## Migration

### Migration from zen-watcher

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

### Migration from shared/security.HardenedHTTPClient

If you're migrating from `zen-platform/src/shared/security`:

**Before:**
```go
import "github.com/kube-zen/zen-platform/src/shared/security"

client := security.NewHardenedHTTPClient(security.DefaultHTTPClientConfig())
resp, err := client.Get(ctx, url)
```

**After (using compatibility alias):**
```go
import sdkhttp "github.com/kube-zen/zen-sdk/pkg/http"

// Option 1: Use compatibility alias (deprecated but works)
client := sdkhttp.NewHardenedHTTPClient(nil)
resp, err := client.Get(ctx, url)

// Option 2: Use new API (recommended)
client := sdkhttp.NewClient(nil)
resp, err := client.Get(ctx, url)
```

**Note:** The `HardenedHTTPClient` type and `NewHardenedHTTPClient` function are available as compatibility aliases but are deprecated. Use `Client` and `NewClient` instead.

## Related

- [zen-sdk/pkg/config](../config/README.md) - Configuration helpers
- [zen-sdk/pkg/logging](../logging/README.md) - Structured logging
- [zen-sdk/pkg/retry](../retry/README.md) - Retry logic (can be combined with HTTP client)

