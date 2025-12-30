# Contributing to Zen SDK

Thank you for your interest in contributing to Zen SDK!

## Philosophy

Zen SDK is a **shared library**, not a monorepo. Keep it focused on cross-cutting concerns:

- ✅ Leader election wrappers
- ✅ Metrics setup
- ✅ Logging configuration
- ✅ Webhook helpers
- ❌ Business logic (belongs in individual tools)
- ❌ Tool-specific code (belongs in that tool's repo)

## Development Setup

```bash
git clone https://github.com/kube-zen/zen-sdk
cd zen-sdk
go mod download
```

## Making Changes

1. Create a feature branch
2. Make your changes
3. Run tests: `make test`
4. Run linters: `make lint`
5. Commit and push

## Adding a New Package

If you want to add a new cross-cutting concern:

1. **Is it truly cross-cutting?**
   - Will at least 3 tools use it?
   - Is it infrastructure, not business logic?
   - If no, it belongs in the tool's repo, not SDK

2. **Create the package:**
   ```bash
   mkdir -p pkg/your-package
   # Add your code
   ```

3. **Add tests:**
   ```bash
   # Create pkg/your-package/your_package_test.go
   ```

4. **Update README.md** with usage examples

5. **Document the API** clearly

## Versioning

Zen SDK follows semantic versioning. Breaking changes require a major version bump.

- **v1.0.0** - Initial stable release
- **v1.1.0** - New features (backward compatible)
- **v2.0.0** - Breaking changes

## Testing

```bash
# Run all tests
make test

# Run specific package tests
go test ./pkg/leader/...
```

## Code Standards

- Use `gofmt` (enforced by `make fmt`)
- Use `go vet` (enforced by `make vet`)
- Write tests for new features
- Document exported functions

## Questions?

Open an issue or check existing documentation.

---

**Remember**: Keep it simple, focused, and reusable.

