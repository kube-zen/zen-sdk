# HTTP Client Migration Status

**Status:** Analysis of components using `HardenedHTTPClient` and migration to `zen-sdk/pkg/http`

## Summary

- **Total Components Analyzed:** 16
- **Using HardenedHTTPClient:** 2 components (zen-platform services, zen-watcher)
- **Should Migrate:** 2 components
- **Using Basic http.Client (should consider):** 3 components (zen-flow, zen-lead)
- **No HTTP Client Usage:** 11 components

## Component Analysis

### ✅ Already Using zen-sdk/pkg/http

**zen-watcher:**
- ✅ **Status:** Already migrated to `zen-sdk/pkg/http` (v0.2.7-alpha)
- **Note:** Has local `pkg/http/client.go` with `HardenedHTTPClient` that wraps zen-sdk Client
- **Action:** Consider removing local wrapper and using zen-sdk directly

### ❌ Should Migrate to zen-sdk/pkg/http

**zen-platform/src/saas/bff:**
- ❌ **Status:** Using `security.NewHardenedHTTPClient` from `zen-platform/src/shared/security`
- **Files:**
  - `src/main.go` (line 932)
  - `src/handlers/me.go` (line 106)
  - `src/handlers/exceptions.go` (line 39)
  - `src/handlers/adapter_registration_proxy.go` (line 31)
  - `src/handlers/clusters.go` (line 56)
  - `src/handlers/observations.go` (line 45)
  - `src/handlers/security_events_dashboard.go` (line 32)
- **Action:** Migrate to `zen-sdk/pkg/http.NewClient()` or `zen-sdk/pkg/http.NewHardenedHTTPClient()` (compatibility alias)

**zen-platform/src/saas/back:**
- ❌ **Status:** Using `security.NewHardenedHTTPClient` from `zen-platform/src/shared/security`
- **Files:**
  - `src/services/email_service.go` (line 84)
  - `src/services/slack_notification_service.go` (line 40)
  - `src/handlers/queue_event_handler.go` (line 41)
  - `src/gitops/client.go` (line 56)
- **Action:** Migrate to `zen-sdk/pkg/http.NewClient()` or `zen-sdk/pkg/http.NewHardenedHTTPClient()` (compatibility alias)

### ⚠️ Should Consider Using zen-sdk/pkg/http

**zen-flow:**
- ⚠️ **Status:** Using basic `http.Client{}` (no retry, no metrics, no connection pooling)
- **Files:**
  - `pkg/controller/artifacts.go` (line 142)
  - `pkg/controller/reconciler_test.go` (line 166)
  - `test/integration/integration_test.go` (line 166)
- **Action:** Consider migrating to `zen-sdk/pkg/http.NewClient()` for better resilience and observability

**zen-lead:**
- ⚠️ **Status:** Using basic `http.Client{Timeout: 10 * time.Second}` (no retry, no metrics)
- **Files:**
  - `test/integration/experimental_features_test.go` (line 89)
- **Action:** Consider migrating to `zen-sdk/pkg/http.NewClient()` for better resilience

### ✅ No HTTP Client Usage (No Action Needed)

- **zen-gc:** No HTTP client usage
- **zen-lock:** No HTTP client usage
- **zen-admin:** No HTTP client usage
- **zen-platform/src/saas/auth:** No HardenedHTTPClient usage
- **zen-platform/src/saas/websocket:** No HardenedHTTPClient usage
- **zen-platform/src/saas/cluster-registry:** No HardenedHTTPClient usage
- **zen-platform/src/saas/bridge:** No HardenedHTTPClient usage
- **zen-platform/src/saas/integrations:** No HardenedHTTPClient usage
- **zen-platform/src/saas/ingester:** No HardenedHTTPClient usage
- **zen-platform/src/saas/gitops:** No HardenedHTTPClient usage (uses gitops client wrapper)

## Migration Priority

### High Priority (Production Services)

1. **zen-platform/src/saas/bff** - Multiple handlers using HardenedHTTPClient
2. **zen-platform/src/saas/back** - Multiple services using HardenedHTTPClient

### Medium Priority (Improve Resilience)

3. **zen-flow** - Controller making HTTP requests without retry/metrics
4. **zen-lead** - Test code using basic HTTP client

## Migration Guide

### From `security.NewHardenedHTTPClient`

**Before:**
```go
import "github.com/kube-zen/zen-platform/src/shared/security"

client := security.NewHardenedHTTPClient(security.DefaultHTTPClientConfig())
resp, err := client.Get(ctx, url)
```

**After (Option 1 - Compatibility Alias):**
```go
import sdkhttp "github.com/kube-zen/zen-sdk/pkg/http"

client := sdkhttp.NewHardenedHTTPClient(nil) // Uses defaults
resp, err := client.Get(ctx, url)
```

**After (Option 2 - Recommended):**
```go
import sdkhttp "github.com/kube-zen/zen-sdk/pkg/http"

client := sdkhttp.NewClient(nil) // Uses defaults
resp, err := client.Get(ctx, url)
```

### From Basic `http.Client`

**Before:**
```go
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Get(url)
```

**After:**
```go
import sdkhttp "github.com/kube-zen/zen-sdk/pkg/http"

config := &sdkhttp.ClientConfig{
    Timeout: 10 * time.Second,
    ServiceName: "zen-flow",
}
client := sdkhttp.NewClient(config)
resp, err := client.Get(ctx, url)
```

## Benefits of Migration

1. **Retry Logic:** Automatic retry with exponential backoff for network errors and 5xx status codes
2. **Prometheus Metrics:** Built-in metrics for requests, retries, errors, and latency
3. **Connection Pooling:** Optimized connection reuse
4. **Rate Limiting:** Optional rate limiting support
5. **Structured Logging:** Integration with zen-sdk logging
6. **Consistency:** Same HTTP client implementation across all OSS components
7. **Middleware Support:** Request/response middleware hooks

## Next Steps

1. ✅ **Completed:** Added HardenedHTTPClient functionality to zen-sdk
2. ⏳ **Pending:** Migrate zen-platform/src/saas/bff
3. ⏳ **Pending:** Migrate zen-platform/src/saas/back
4. ⏳ **Optional:** Migrate zen-flow and zen-lead for better resilience

