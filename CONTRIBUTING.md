# Contributing to jira-beads-sync

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to this project.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`) version 3.x or later
- `protoc-gen-go` plugin
- `golangci-lint` for code linting
- Make (optional, but recommended)

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/conallob/jira-beads-sync.git
   cd jira-beads-sync
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   ```

3. **Generate protobuf files:**
   ```bash
   make proto
   ```

4. **Build the project:**
   ```bash
   make build
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**, following the coding standards below

3. **Generate protobuf files if you modified `.proto` files:**
   ```bash
   make proto
   ```

4. **Run tests:**
   ```bash
   make test
   ```

5. **Format and lint your code:**
   ```bash
   make fmt
   make lint
   ```

6. **Run all verification steps:**
   ```bash
   make verify
   ```

### Coding Standards

- **Follow Go conventions**: Use `gofmt` and `golangci-lint`
- **Write tests**: All new functionality should have corresponding tests
- **Use standard libraries**: Prefer standard library over external dependencies
- **Avoid single-letter variables**: Except for trivial loop indices
- **Document exported functions**: Add godoc comments for all exported functions

### Protocol Buffers

When modifying data structures:

1. **Update `.proto` files first** in the `proto/` directory
2. **Regenerate Go code** using `make proto`
3. **Update rendering layers** if field names or types change
4. **Update tests** to reflect schema changes

The `.proto` files are the source of truth for all data structures.

### Testing

- Write unit tests for all new functionality
- Ensure tests cover edge cases and error conditions
- Run tests with race detection: `go test -race ./...`
- Maintain or improve test coverage

### Commit Messages

Use conventional commit format:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test changes
- `refactor:` for code refactoring
- `ci:` for CI/CD changes
- `chore:` for maintenance tasks

Example:
```
feat: add support for Jira custom fields
fix: handle nil assignee in converter
docs: update installation instructions
```

## Pull Request Process

1. **Update documentation** if you're changing functionality
2. **Add tests** for new functionality
3. **Ensure all tests pass**: `make verify`
4. **Update CHANGELOG.md** if applicable
5. **Submit your PR** with a clear description of changes

### PR Requirements

- All tests must pass
- Code must be formatted (`make fmt`)
- Linter must pass with no errors (`make lint`)
- Commit messages follow conventional format
- Documentation is updated if needed

## CI/CD

This project uses GitHub Actions for CI/CD:

- **Test workflow** (`test.yml`): Runs on all pushes and PRs
  - Tests on Go 1.21, 1.22, and 1.23
  - Runs linter and formatting checks
  - Verifies protobuf files are up to date

- **Release workflow** (`release.yml`): Runs on version tags
  - Builds binaries for all platforms
  - Creates GitHub releases
  - Publishes Docker images
  - Updates Homebrew formula

## Reporting Issues

When reporting issues, please include:

- Go version (`go version`)
- Operating system and architecture
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Relevant logs or error messages

## Questions?

Feel free to open an issue for questions or discussions about the project.

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (Apache 2.0).
