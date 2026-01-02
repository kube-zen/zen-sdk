# zenctl Implementation Report

## Overview

`zenctl` is an operator-grade CLI tool for inspecting Zen Kubernetes resources (DeliveryFlows, Destinations, Ingesters). This report summarizes the implementation completed for P701-P704.

## New Commands

### 1. `zenctl status`

Summarizes DeliveryFlows, Destinations, and Ingesters across namespaces.

**Example:**
```bash
zenctl status -A
```

**Output:**
- DeliveryFlows: namespace, name, active target, entitlement, ready, age
- Destinations: namespace, name, type, transport, health, ready, age
- Ingesters: namespace, name, sources, health, last seen, entitled, blocked, ready, age

### 2. `zenctl flows`

Lists DeliveryFlows in table format aligned with ACTIVE_TARGET_UX_GUIDE.md.

**Example:**
```bash
zenctl flows -A
```

**Columns:**
- NAMESPACE | NAME | ACTIVE_TARGET | ENTITLEMENT | ENTITLEMENT_REASON | READY | AGE

### 3. `zenctl explain flow <name>`

Prints detailed information about a specific DeliveryFlow.

**Example:**
```bash
zenctl explain flow my-flow -n zen-apps
```

**Output includes:**
- Resolved sourceKey list from `spec.sources`
- Outputs with active target per output
- Last failover timestamp/reason (if present)
- Entitlement condition + reason

### 4. `zenctl doctor`

Runs diagnostic checks for common misconfigurations (no network dependency).

**Example:**
```bash
zenctl doctor
```

**Checks:**
- CRDs installed: DeliveryFlow, Destination, Ingester
- Controllers present: zen-ingester, zen-watcher deployments (best-effort)
- Status subresources exist on CRDs (best-effort via discovery)

**Exit codes:**
- `0`: All checks PASS
- `1`: One or more checks FAIL

## Dynamic CRD Discovery

`zenctl` uses dynamic client + discovery (no hardcoded GVRs):

1. **Discovery Client**: Uses `k8s.io/client-go/discovery/cached/disk` for efficient API discovery
2. **RESTMapper**: Uses `restmapper.NewDeferredDiscoveryRESTMapper` to resolve GVK → GVR
3. **CRD Resolution**: Looks up resources by GVK:
   - `DeliveryFlow`: `routing.zen.kube-zen.io/v1alpha1`
   - `Destination`: `routing.zen.kube-zen.io/v1alpha1`
   - `Ingester`: `zen.kube-zen.io/v1alpha1`

4. **Error Handling**: If a CRD is missing, prints precise actionable message:
   ```
   DeliveryFlow CRD not installed; enable crds.enabled or apply CRDs separately.
   ```
   Exits non-zero on critical errors.

## Assumptions About Group/Version/Kind Names

The implementation assumes the following GVKs (defined in `cmd/zenctl/internal/discovery/discovery.go`):

- **DeliveryFlow**: `routing.zen.kube-zen.io/v1alpha1/DeliveryFlow`
- **Destination**: `routing.zen.kube-zen.io/v1alpha1/Destination`
- **Ingester**: `zen.kube-zen.io/v1alpha1/Ingester`

These are resolved dynamically via discovery, so if the CRDs exist with different groups/versions, the tool will discover them. However, the tool specifically looks for these expected GVKs.

## Key Implementation Details

### Kubeconfig Resolution

Supports multiple kubeconfig sources (in order):
1. In-cluster config (when running inside Kubernetes)
2. `$KUBECONFIG` environment variable
3. Default `~/.kube/config`

### Global Flags

- `--context`: Kubernetes context to use
- `--namespace, -n`: Kubernetes namespace
- `--all-namespaces, -A`: List resources across all namespaces
- `--kubeconfig`: Path to kubeconfig file

### Output Formats

All commands support:
- `table` (default): Human-readable tables
- `json`: JSON output (raw Kubernetes objects)
- `yaml`: YAML output (raw Kubernetes objects)

### Security

**Important:** `zenctl` never prints secrets or sensitive data. Only status fields and non-sensitive spec fields are displayed. The tool does not dump full spec objects that may contain credentials.

### Status Field Parsing

Uses `unstructured.Unstructured` parsing to read:
- `status.outputs[]`
- `status.conditions[]`
- `spec.outputs[]` (for sourceKey resolution in explain command)

### Human-Readable Formatting

- **Active Target**: Shows as `namespace/name` when both present, otherwise just `name`
- **Entitlement**: Human-readable labels:
  - `True` + `<none>` → "Entitled"
  - `False` + `GracePeriod` → "Grace Period"
  - `False` + `Expired` → "Expired"
  - `False` + `NotEntitled` → "Not Entitled"
  - `Unknown` → "Unknown"
- **Age**: Relative time formatting (e.g., "5m", "2h", "3d")

## Files Created

### Core Implementation

- `cmd/zenctl/main.go` - Main entrypoint with cobra CLI structure
- `cmd/zenctl/internal/client/client.go` - Kubeconfig client setup
- `cmd/zenctl/internal/discovery/discovery.go` - CRD discovery and GVR resolution
- `cmd/zenctl/internal/output/output.go` - Output formatting helpers
- `cmd/zenctl/internal/resources/resources.go` - Resource parsing and listing
- `cmd/zenctl/internal/commands/options.go` - Global options context
- `cmd/zenctl/internal/commands/status.go` - Status command implementation
- `cmd/zenctl/internal/commands/flows.go` - Flows command implementation
- `cmd/zenctl/internal/commands/explain.go` - Explain command implementation
- `cmd/zenctl/internal/commands/doctor.go` - Doctor command implementation

### Tests

- `cmd/zenctl/internal/output/output_test.go` - Formatting helper tests
- `cmd/zenctl/internal/discovery/discovery_test.go` - Discovery tests

### Documentation

- `docs/zenctl/README.md` - Complete CLI documentation
- `API_REFERENCE.md` - Updated to mention zenctl

### Build

- `Makefile` - Added `make zenctl` target

## Dependencies Added

- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML output support (already in go.mod)

## Testing

### Unit Tests

- Formatting helpers: `FormatEntitlement`, `FormatActiveTarget`, `FormatDuration`, `ParseFormat`
- Discovery: Expected GVKs validation

### Build

```bash
make zenctl
```

Builds the binary to `./zenctl`.

## Acceptance Criteria

✅ **P701**: Created zenctl operator CLI with:
- ✅ kubeconfig resolution (in-cluster, $KUBECONFIG, ~/.kube/config)
- ✅ Global flags (--context, --namespace/-n, --all-namespaces/-A)
- ✅ Subcommands: status, flows, explain flow
- ✅ Dynamic client + discovery (no hardcoded GVRs)
- ✅ Output formats: table, json, yaml
- ✅ No secrets printed

✅ **P702**: Added doctor command:
- ✅ Checks CRDs installed
- ✅ Checks controllers present (best-effort)
- ✅ Checks status subresources (best-effort)
- ✅ Reports PASS/WARN/FAIL with remediations
- ✅ Exits 0 if all PASS, 1 if any FAIL

✅ **P703**: Documentation:
- ✅ `docs/zenctl/README.md` with install/build instructions and examples
- ✅ `API_REFERENCE.md` updated to mention zenctl

✅ **P704**: CI sanity:
- ✅ Tests for formatting helpers
- ✅ Tests for discovery logic
- ✅ `make zenctl` target in Makefile

## Usage Examples

```bash
# Quick status check
zenctl status -A

# List all flows
zenctl flows -A

# Explain a flow
zenctl explain flow my-flow -n production

# Run diagnostics
zenctl doctor

# JSON output
zenctl flows -o json | jq '.[] | .metadata.name'
```

## Next Steps (Optional Enhancements)

- Add more detailed error messages with remediation hints
- Add filtering options (e.g., `zenctl flows --ready`)
- Add watch mode for real-time updates
- Add completion support (bash/zsh)
- Add integration tests against a test cluster

