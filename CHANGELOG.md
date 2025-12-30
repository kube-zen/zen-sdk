# Changelog

All notable changes to zen-sdk will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0-alpha] - 2025-01-XX

### 🎉 Initial Alpha Release

**First release** of Zen SDK as a shared library for cross-cutting concerns.

#### Added

**Core Packages:**
- **pkg/leader**: Wrapper around controller-runtime leader election
  - Simple Options API
  - Consistent configuration across tools
  - ManagerOptions helper function
  
- **pkg/metrics**: Prometheus metrics recorder
  - Standard reconciliation metrics
  - Error tracking
  - Component-specific labels
  - Idempotent registration

- **pkg/logging**: Structured logging
  - Zap-based structured logging
  - Component name context
  - Development mode detection
  - Consistent format

- **pkg/webhook**: Webhook helpers
  - JSON patch generation
  - TLS secret validation
  - NamespacedName extraction
  - Kubernetes patch utilities

**Documentation:**
- README with usage examples
- Architecture documentation
- Migration guide for existing tools
- Contributing guidelines
- Project summary

**Examples:**
- Leader election example
- Metrics usage example
- Logging usage example
- Webhook usage example

**Infrastructure:**
- Go module setup
- Makefile with build targets
- Tests for all packages
- Git repository initialized

#### Known Limitations

- Logging package uses direct zap logger (not controller-runtime logger wrapper)
- Metrics registration is idempotent but may create duplicate metrics if called multiple times with same component name
- Webhook package requires unstructured conversion for some operations

---

## Roadmap

### v0.2.0 (Planned)
- [ ] Enhanced logging integration with controller-runtime
- [ ] Metrics registry per component to avoid duplicates
- [ ] Additional webhook utilities
- [ ] Health check helpers (pkg/health)

### v0.3.0 (Planned)
- [ ] Tracing support (pkg/tracing)
- [ ] Configuration management (pkg/config)
- [ ] Client helpers (pkg/client)

### v1.0.0 (Future)
- [ ] Stable API
- [ ] Full documentation
- [ ] Production-ready

---

**Current Version:** 0.1.0-alpha  
**Next Milestone:** v0.2.0 - Enhanced features

