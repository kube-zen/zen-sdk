# Zen SDK Documentation

Welcome to the Zen SDK documentation. Zen SDK is a shared library providing reusable components for Kubernetes operators and controllers, eliminating code duplication across zen-flow, zen-lock, zen-watcher, and other Zen tools.

## Getting Started

- [Quick Start](../README.md#quick-start) - Get started in 5 minutes
- [Installation](../README.md#quick-start) - Install zen-sdk in your project
- [Examples](../README.md#examples) - Code examples for all components

## Core Components

- **Leader Election** - Consistent leader election across controllers
- **Metrics** - Prometheus metrics recording
- **Logging** - Structured logging with context
- **Webhook** - Admission webhook utilities
- **Health Checks** - Health check interfaces
- **Lifecycle** - Graceful shutdown handling
- **Configuration** - Configuration validation
- **Retry** - Retry logic with backoff
- **Kubernetes Utilities** - Name generation, validation, and more

See the [README](../README.md#components) for detailed component documentation.

## Development

- [Development Guide](DEVELOPMENT.md) - Setup and development workflows
- [Release Process](RELEASE.md) - How releases are made
- [OSS Boundary](OSS_BOUNDARY.md) - Open source boundary guidelines
- [Leadership Contract](LEADERSHIP_CONTRACT.md) - Leader election contract

## Migration

- [Migration Guide](../MIGRATION_GUIDE.md) - Migrate existing tools to zen-sdk
- [Migration Examples](../MIGRATION_EXAMPLES.md) - Practical migration examples

## Resources

- [GitHub Repository](https://github.com/kube-zen/zen-sdk)
- [API Reference](../API_REFERENCE.md) - Complete API documentation
- [Architecture](../ARCHITECTURE.md) - Design and architecture
- [Contributing](../CONTRIBUTING.md) - Contribution guidelines
- [Security Policy](../SECURITY.md) - Security reporting
- [Code of Conduct](../CODE_OF_CONDUCT.md) - Community standards
- [Changelog](../CHANGELOG.md) - Version history
