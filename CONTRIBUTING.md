# Contributing to Forge

Thank you for your interest in contributing to Forge! This document provides guidelines and information for contributors.

## 🚨 Alpha Software Notice

Forge is currently in **alpha stage**. Many features use mock implementations for testing purposes. Please check our [implementation status](README.md#implementation-status) before contributing.

## Getting Started

### Prerequisites
- Go 1.21 or later
- Make
- Git

### Development Setup
```bash
# Clone the repository
git clone https://github.com/ataiva-software/forge.git
cd chisel

# Install dependencies and build
make build

# Run tests
make test

# Run integration tests
make test-integration
```

## Development Workflow

### 1. Test-Driven Development (TDD)
We use TDD for all new features:

1. **Write failing tests first**
2. **Implement minimal code to pass**
3. **Refactor and improve**
4. **Repeat**

Example:
```bash
# Create test file
touch pkg/newfeature/feature_test.go

# Write failing tests
# Implement feature
# Run tests
go test ./pkg/newfeature -v
```

### 2. Code Organization
- `pkg/` - Core packages and libraries
- `cmd/` - CLI commands and main entry points
- `examples/` - Example configurations and use cases
- `docs/` - Documentation (auto-generated)

### 3. Testing Standards
- **Unit tests** for all packages
- **Integration tests** for end-to-end workflows
- **Mock implementations** for external dependencies
- **100% test coverage** goal for new code

## Contribution Types

### 🐛 Bug Fixes
1. Create an issue describing the bug
2. Write a test that reproduces the bug
3. Fix the bug
4. Ensure all tests pass

### ✨ New Features
1. Check the [roadmap](README.md#implementation-status) first
2. Create a feature request issue
3. Discuss the approach with maintainers
4. Implement using TDD
5. Update documentation

### 📚 Documentation
1. Update relevant README files
2. Add examples for new features
3. Update API documentation
4. Regenerate docs with `make docs`

### 🧪 Testing
1. Add test coverage for untested code
2. Improve existing tests
3. Add integration tests
4. Performance testing

## Code Standards

### Go Style
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` and `golint`
- Write clear, self-documenting code
- Add comments for complex logic

### Testing Style
```go
func TestFeature_Scenario(t *testing.T) {
    // Arrange
    input := setupTestData()
    
    // Act
    result, err := feature.Process(input)
    
    // Assert
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Commit Messages
Use conventional commits:
```
type(scope): description

feat(providers): add kubernetes provider
fix(ssh): handle connection timeouts
docs(readme): update installation instructions
test(core): add module validation tests
```

## Pull Request Process

### 1. Before Submitting
- [ ] Tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] Self-review completed

### 2. PR Requirements
- Clear description of changes
- Link to related issues
- Test coverage for new code
- Documentation updates
- No breaking changes (unless discussed)

### 3. Review Process
1. Automated tests run
2. Code review by maintainers
3. Address feedback
4. Approval and merge

## Priority Areas for Contribution

### High Priority
1. **Real provider implementations** - Replace mock implementations
2. **SSH/WinRM connections** - Production-ready remote execution
3. **Error handling** - Robust error handling and recovery
4. **Performance** - Optimize execution speed

### Medium Priority
1. **Additional providers** - Database, network, cloud resources
2. **Cloud integrations** - GCP, DigitalOcean, etc.
3. **Monitoring** - Enhanced metrics and observability
4. **Security** - Vulnerability scanning, secure defaults

### Nice to Have
1. **Web UI enhancements** - Better dashboard features
2. **Plugin system** - WASM-based extensions
3. **Advanced workflows** - Complex orchestration patterns
4. **Performance optimizations** - Caching, parallelization

## Getting Help

### Communication Channels
- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - General questions and ideas
- **Pull Requests** - Code review and collaboration

### Resources
- [Architecture Overview](docs/architecture.md)
- [API Documentation](docs/api.md)
- [Examples](examples/)
- [Roadmap](README.md#implementation-status)

## Recognition

Contributors will be:
- Listed in the project README
- Mentioned in release notes
- Invited to join the maintainer team (for significant contributions)

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.

## License

By contributing to Forge, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to Forge!**

Your contributions help make infrastructure automation better for everyone.
