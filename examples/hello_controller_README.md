# Hello Controller - Minimal Controller Example

This is a minimal, copy-paste ready controller example that demonstrates basic controller-runtime usage with zen-sdk packages.

## Features

- ✅ **Minimal dependencies**: Only controller-runtime and zen-sdk
- ✅ **Copy-paste ready**: Single file, no external setup required
- ✅ **Fast compilation**: Compiles in under 2 minutes
- ✅ **Production-ready patterns**: Uses zen-sdk for logging, metrics, lifecycle, and health checks
- ✅ **Graceful shutdown**: Proper signal handling and cleanup

## What It Does

The controller watches ConfigMaps in all namespaces and logs when they are:
- Created
- Updated
- Deleted

## Quick Start

### 1. Copy the Example

```bash
# Copy the example file
cp examples/hello_controller.go my-controller/main.go
cd my-controller
```

### 2. Initialize Go Module

```bash
go mod init hello-controller
```

### 3. Install Dependencies

```bash
go get github.com/kube-zen/zen-sdk@latest
go get sigs.k8s.io/controller-runtime@latest
go get k8s.io/api@latest
go get k8s.io/apimachinery@latest
go get k8s.io/client-go@latest
```

### 4. Run the Controller

```bash
# Make sure you have a valid kubeconfig
export KUBECONFIG=~/.kube/config

# Run the controller
go run main.go
```

## What You'll See

```
{"level":"info","ts":"2024-01-01T12:00:00Z","msg":"Starting Hello Controller","component":"hello-controller"}
{"level":"info","ts":"2024-01-01T12:00:01Z","msg":"Starting manager","component":"hello-controller"}
{"level":"info","ts":"2024-01-01T12:00:02Z","msg":"ConfigMap reconciled","component":"hello-controller","namespace":"default","name":"my-config","data_keys":2,"resource_version":"12345"}
```

## Testing

### Create a ConfigMap

```bash
kubectl create configmap test-config --from-literal=key1=value1 --from-literal=key2=value2
```

You should see a log entry in the controller output.

### Check Health Endpoints

```bash
# Health check
curl http://localhost:8081/healthz

# Readiness check
curl http://localhost:8081/readyz

# Metrics
curl http://localhost:8080/metrics
```

## Customization

### Watch a Specific Namespace

Change the manager options:

```go
mgr, err := ctrl.NewManager(cfg, ctrl.Options{
    Scheme: scheme,
    Namespace: "my-namespace", // Watch only this namespace
    MetricsBindAddress: ":8080",
    HealthProbeBindAddress: ":8081",
})
```

### Watch a Different Resource

Change the controller setup:

```go
// Watch Pods instead of ConfigMaps
if err := ctrl.NewControllerManagedBy(mgr).
    For(&corev1.Pod{}).
    Complete(reconciler); err != nil {
    // ...
}
```

### Add Leader Election

```go
import "github.com/kube-zen/zen-sdk/pkg/leader"

// In main()
leaderOpts := leader.Options{
    LeaseName: "hello-controller-leader-election",
    Enable:    true,
    Namespace: "default",
}

mgr, err := ctrl.NewManager(cfg, ctrl.Options{
    Scheme: scheme,
}, leader.Setup(leaderOpts))
```

## Next Steps

- Read the [zen-sdk README](../../README.md) for more features
- Check [QUICKSTART.md](../../QUICKSTART.md) for more examples
- See other examples in [examples/](../examples/) directory

## Troubleshooting

### "unable to get rest config"

Make sure you have a valid kubeconfig:
```bash
kubectl cluster-info
```

### "context deadline exceeded"

The controller is trying to connect to the cluster. Check:
- Cluster is accessible
- KUBECONFIG is set correctly
- You have permissions to watch ConfigMaps

### Port already in use

Change the port in the manager options:
```go
MetricsBindAddress: ":8082",  // Use different port
HealthProbeBindAddress: ":8083",
```

