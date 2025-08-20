# Contributing to Chisel

Thank you for your interest in contributing to Chisel! This document provides guidelines and information for contributors.

## Development Philosophy

Chisel follows Test-Driven Development (TDD) principles:

1. **Write failing tests first** - Before implementing any feature, write tests that describe the expected behavior
2. **Implement minimal code to pass** - Write just enough code to make the tests pass
3. **Refactor and improve** - Clean up the code while keeping tests green
4. **Repeat** - Continue this cycle for all new features

## Getting Started

### Prerequisites

- Go 1.21 or later
- Make
- Git

### Setting up the development environment

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/chisel.git
   cd chisel
   ```
3. Install dependencies:
   ```bash
   make deps
   ```
4. Run tests to ensure everything works:
   ```bash
   make test
   ```

## Development Workflow

### 1. Create a feature branch
```bash
git checkout -b feature/your-feature-name
```

### 2. Write tests first
Before implementing any functionality, write comprehensive tests:

```bash
# Create test file
touch pkg/your-package/your-feature_test.go

# Write failing tests that describe the expected behavior
# Run tests to confirm they fail
make test
```

### 3. Implement the feature
Write the minimal code needed to make your tests pass:

```bash
# Implement your feature
# Run tests frequently
make test
```

### 4. Refactor and improve
Once tests are passing, improve the code quality:

- Remove duplication
- Improve naming
- Add documentation
- Ensure error handling is robust

### 5. Run the full test suite
```bash
make test
make test-integration  # When available
make lint
```

## Code Standards

### Go Code Style
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Write clear, concise comments
- Handle errors appropriately
- Use interfaces for testability

### Testing Standards
- Write table-driven tests when appropriate
- Use descriptive test names
- Test both happy path and error cases
- Mock external dependencies
- Aim for high test coverage

### Example Test Structure
```go
func TestFeature_Method(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        {
            name:    "invalid input",
            input:   invalidInput,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Method(tt.input)
            if tt.wantErr {
                if err == nil {
                    t.Error("expected error but got none")
                }
                return
            }
            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Project Structure

```
chisel/
├── cmd/chisel/          # Main CLI application
├── pkg/
│   ├── cli/            # CLI commands and UI
│   ├── core/           # Core business logic
│   ├── providers/      # Resource providers
│   ├── inventory/      # Inventory management
│   ├── ssh/            # SSH connection handling
│   └── types/          # Core types and interfaces
├── internal/           # Private application code
├── tests/
│   ├── unit/          # Unit tests
│   └── integration/   # Integration tests
├── examples/          # Example configurations
└── docs/             # Documentation
```

## Submitting Changes

### Pull Request Process

1. Ensure all tests pass
2. Update documentation if needed
3. Add entries to CHANGELOG.md for user-facing changes
4. Create a pull request with:
   - Clear title and description
   - Reference to any related issues
   - Screenshots/examples if applicable

### Pull Request Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] New tests added for new functionality
- [ ] Integration tests pass (if applicable)

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (if user-facing)
```

## Reporting Issues

### Bug Reports
When reporting bugs, please include:
- Chisel version
- Operating system and version
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs or error messages

### Feature Requests
For feature requests, please provide:
- Clear description of the feature
- Use case and motivation
- Proposed implementation approach (if any)
- Willingness to contribute the implementation

## Community Guidelines

- Be respectful and inclusive
- Help others learn and grow
- Focus on constructive feedback
- Follow the project's code of conduct

## Getting Help

- Check existing issues and documentation
- Ask questions in GitHub Discussions
- Join our community channels (when available)

## Recognition

Contributors will be recognized in:
- CHANGELOG.md for significant contributions
- README.md contributors section
- Release notes for major features

Thank you for contributing to Chisel! 🪓
