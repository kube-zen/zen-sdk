# Lifecycle Package

**Graceful shutdown helpers for Zen components**

This package provides standardized graceful shutdown functionality for all Zen components, using Go's standard `signal.NotifyContext()` (Go 1.16+).

## Features

- ✅ Signal handling (`SIGINT`, `SIGTERM`)
- ✅ HTTP server graceful shutdown
- ✅ gRPC server graceful shutdown
- ✅ Worker service shutdown coordination
- ✅ Structured logging integration
- ✅ Configurable timeouts

## Quick Start

### HTTP Server Component

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/kube-zen/zen-sdk/pkg/lifecycle"
    "github.com/kube-zen/zen-sdk/pkg/logging"
)

func main() {
    ctx := context.Background()
    logger := logging.NewLogger("zen-back")
    
    // Create shutdown context
    shutdownCtx, cancel := lifecycle.ShutdownContext(ctx, "zen-back")
    defer cancel()
    
    // Create HTTP server
    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }
    
    // Start server in goroutine
    go func() {
        logger.Info("Starting server", logging.String("addr", ":8080"))
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error(err, "Server failed", logging.Operation("startup"))
        }
    }()
    
    // Wait for shutdown signal
    <-shutdownCtx.Done()
    
    // Graceful shutdown
    if err := lifecycle.ShutdownHTTPServer(shutdownCtx, server, "zen-back", 30*time.Second); err != nil {
        logger.Error(err, "Shutdown failed", logging.Operation("shutdown"))
    }
}
```

### Worker Service Component

```go
package main

import (
    "context"
    "sync"
    "time"

    "github.com/kube-zen/zen-sdk/pkg/lifecycle"
    "github.com/kube-zen/zen-sdk/pkg/logging"
)

func main() {
    ctx := context.Background()
    logger := logging.NewLogger("zen-back-workers")
    
    // Create shutdown context
    shutdownCtx, cancel := lifecycle.ShutdownContext(ctx, "zen-back-workers")
    defer cancel()
    
    // Start workers
    var wg sync.WaitGroup
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        runWorker(shutdownCtx)
    }()
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        runAnotherWorker(shutdownCtx)
    }()
    
    // Wait for shutdown signal
    <-shutdownCtx.Done()
    
    logger.Info("Shutting down workers...", logging.Operation("shutdown"))
    
    // Wait for all workers to finish
    wg.Wait()
    
    logger.Info("Workers stopped", logging.Operation("shutdown_complete"))
}

func runWorker(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Do work
        }
    }
}
```

### gRPC Server Component

```go
package main

import (
    "context"
    "net"
    "time"

    "github.com/kube-zen/zen-sdk/pkg/lifecycle"
    "google.golang.org/grpc"
)

func main() {
    ctx := context.Background()
    
    // Create shutdown context
    shutdownCtx, cancel := lifecycle.ShutdownContext(ctx, "zen-bridge")
    defer cancel()
    
    // Create gRPC server
    lis, err := net.Listen("tcp", ":9090")
    if err != nil {
        panic(err)
    }
    
    grpcServer := grpc.NewServer()
    // ... register services ...
    
    // Start server in goroutine
    go func() {
        if err := grpcServer.Serve(lis); err != nil {
            // Handle error
        }
    }()
    
    // Wait for shutdown signal
    <-shutdownCtx.Done()
    
    // Graceful shutdown
    lifecycle.ShutdownGRPCServer(grpcServer, "zen-bridge")
}
```

### Using WaitForShutdown Helper

```go
package main

import (
    "context"
    "sync"

    "github.com/kube-zen/zen-sdk/pkg/lifecycle"
)

func main() {
    ctx := context.Background()
    
    shutdownCtx, cancel := lifecycle.ShutdownContext(ctx, "my-component")
    defer cancel()
    
    var wg sync.WaitGroup
    
    // Start workers
    wg.Add(1)
    go func() {
        defer wg.Done()
        runWorker(shutdownCtx)
    }()
    
    // Wait for shutdown with cleanup
    lifecycle.WaitForShutdown(shutdownCtx, "my-component", func() {
        // Wait for all workers
        wg.Wait()
        
        // Additional cleanup
        cleanup()
    })
}
```

## API Reference

### ShutdownContext

Creates a context that cancels on `SIGINT` or `SIGTERM`.

```go
func ShutdownContext(ctx context.Context, component string) (context.Context, context.CancelFunc)
```

**Parameters:**
- `ctx`: Base context
- `component`: Component name for logging

**Returns:**
- Context that cancels on shutdown signals
- Cancel function

### ShutdownHTTPServer

Gracefully shuts down an HTTP server with timeout.

```go
func ShutdownHTTPServer(ctx context.Context, server *http.Server, component string, timeout time.Duration) error
```

**Parameters:**
- `ctx`: Context (typically from `ShutdownContext`)
- `server`: HTTP server to shutdown
- `component`: Component name for logging
- `timeout`: Shutdown timeout (0 = use default 30s)

**Returns:**
- Error if shutdown fails or times out

### ShutdownGRPCServer

Gracefully shuts down a gRPC server.

```go
func ShutdownGRPCServer(server GRPCServer, component string)
```

**Parameters:**
- `server`: gRPC server implementing `GracefulStop()`
- `component`: Component name for logging

**Note:** `GracefulStop()` blocks until all RPCs finish.

### WaitForShutdown

Waits for context cancellation and runs optional cleanup.

```go
func WaitForShutdown(ctx context.Context, component string, cleanup func())
```

**Parameters:**
- `ctx`: Context (typically from `ShutdownContext`)
- `component`: Component name for logging
- `cleanup`: Optional cleanup function (can be nil)

## Constants

- `DefaultShutdownTimeout`: Default timeout for HTTP server shutdown (30 seconds)

## Best Practices

1. **Always use ShutdownContext**: Create shutdown context early in `main()`
2. **Propagate context**: Pass shutdown context to all goroutines
3. **Respect cancellation**: Check `ctx.Done()` in long-running operations
4. **Set appropriate timeouts**: Use 30s for HTTP servers, adjust for your needs
5. **Log shutdown events**: The package logs automatically, but add component-specific logs if needed

## Migration from Manual Signal Handling

### Before

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := server.Shutdown(shutdownCtx); err != nil {
    // Handle error
}
```

### After

```go
shutdownCtx, cancel := lifecycle.ShutdownContext(context.Background(), "my-component")
defer cancel()

<-shutdownCtx.Done()

if err := lifecycle.ShutdownHTTPServer(shutdownCtx, server, "my-component", 30*time.Second); err != nil {
    // Handle error
}
```

## Why signal.NotifyContext()?

- ✅ **Standard library**: No external dependencies
- ✅ **Modern idiom**: Recommended by Go team (Go 1.16+)
- ✅ **Simple API**: Returns context directly
- ✅ **Works everywhere**: HTTP servers, workers, gRPC, controllers

## Controller Components

For Kubernetes controllers using controller-runtime, use `ctrl.SetupSignalHandler()`:

```go
import ctrl "sigs.k8s.io/controller-runtime"

func main() {
    ctx := ctrl.SetupSignalHandler()
    mgr, err := ctrl.NewManager(cfg, opts)
    if err := mgr.Start(ctx); err != nil {
        // Handle error
    }
}
```

This package is for non-controller components (HTTP servers, workers, etc.).

