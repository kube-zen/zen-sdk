# OSS Boundary Policy

This document defines the boundary between OSS and internal/proprietary components in the Zen project.

## zen-sdk: OSS Library Only

`zen-sdk` is a **pure OSS SDK library** containing only:
- Client-side SDK packages (logging, metrics, leader election, etc.)
- OSS Kubernetes resource types and utilities
- No operator CLIs
- No SaaS integrations
- No proprietary endpoints or authentication flows
- **No committed binaries** (zen-sdk must not ship or commit any CLI binaries)

## zenctl-oss: OSS Operator CLI

The OSS operator CLI (`zenctl-oss`) lives in **`zen-watcher`** repository:
- **Location:** `zen-watcher/cmd/zenctl/`
- **Commands:** status, flows, explain, doctor, adapters, e2e (K8s-only)
- **Build:** `make -C zen-watcher zenctl` (uses GOWORK=off by default)
- **Build with workspace:** `make -C zen-watcher zenctl-workspace`
- **Release:** `make -C zen-watcher release-zenctl` (produces multi-arch binaries + SHA256SUMS)
- **Documentation:** `zen-watcher/docs/zenctl/README.md`

**OSS Scope:**
- Kubernetes cluster operations only
- CRD inspection and status reporting
- No SaaS API endpoints
- No tenant/entitlement SaaS integrations

**Installation:**
1. Download from releases: `zenctl-linux-amd64`, `zenctl-darwin-amd64`, etc.
2. Verify checksums: `sha256sum -c SHA256SUMS`
3. Rename and install: `mv zenctl-linux-amd64 /usr/local/bin/zenctl && chmod +x /usr/local/bin/zenctl`

## zenctl-pro: Internal Operations CLI

The internal/proprietary operations CLI (`zenctl-pro`) lives in:
- **Canonical Location:** `zen-admin/cmd/zenctl-pro/`
- **Commands:** audit (SaaS-only operations)
- **Build:** `go build ./cmd/zenctl-pro`
- **Not distributed as OSS**

## OSS Boundary Enforcement

OSS repositories (`zen-sdk`, `zen-watcher`) must not contain:
- References to `ZEN_API_BASE_URL`
- SaaS API endpoints (e.g., `/v1/audit`, `/v1/clusters`, `/v1/adapters`, `/v1/tenants`)
- SaaS authentication flows
- Tenant/entitlement SaaS handlers
- Internal platform code paths (`src/saas/`, etc.)
- Redis/Cockroach client usage in CLI code paths
- Imports from SaaS-only packages

**Automated Enforcement:**
- Run `bash scripts/test/oss-boundary-gate.sh` to check for violations
- The gate fails on any of the above patterns
- CI should run this check before allowing merges

