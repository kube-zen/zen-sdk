# OSS Boundary Policy

This document defines the boundary between OSS and internal/proprietary components in the Zen project.

## zen-sdk: OSS Library Only

`zen-sdk` is a **pure OSS SDK library** containing only:
- Client-side SDK packages (logging, metrics, leader election, etc.)
- OSS Kubernetes resource types and utilities
- No operator CLIs
- No SaaS integrations
- No proprietary endpoints or authentication flows

## zenctl-oss: OSS Operator CLI

The OSS operator CLI (`zenctl-oss`) lives in **`zen-watcher`** repository:
- Location: `zen-watcher/cmd/zenctl/`
- Commands: status, flows, explain, doctor, adapters, e2e (K8s-only)
- Build: `make -C zen-watcher zenctl`
- Documentation: `zen-watcher/docs/zenctl/README.md`

**OSS Scope:**
- Kubernetes cluster operations only
- CRD inspection and status reporting
- No SaaS API endpoints
- No tenant/entitlement SaaS integrations

## zenctl-pro: Internal Operations CLI

The internal/proprietary operations CLI (`zenctl-pro`) lives in internal repositories:
- Location: `zen-admin/cmd/zenctl-pro` or `zen-platform/src/saas/admin/cmd/zenctl-pro`
- Commands: audit, entitlement management, SaaS-specific operations
- **Not distributed as OSS**

## OSS Boundary Enforcement

OSS repositories (`zen-sdk`, `zen-watcher`) must not contain:
- References to `ZEN_API_BASE_URL`
- SaaS API endpoints (e.g., `/v1/audit`)
- SaaS authentication flows
- Tenant/entitlement SaaS handlers
- Internal platform code paths (`src/saas/`, etc.)

See `scripts/test/oss-boundary-gate.sh` for automated enforcement.

