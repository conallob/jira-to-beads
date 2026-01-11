# Contributing to jira-beads-sync

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to this project.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`) version 33.2 (to ensure consistency with CI)
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

- **Release workflow** (`release.yml`): Runs on version tags (e.g., `v1.0.0`)
  - Builds binaries for all platforms (Linux, macOS, Windows)
  - Creates RPM packages (RHEL, Fedora, CentOS)
  - Creates DEB packages (Debian, Ubuntu)
  - Publishes multi-arch Docker images to GitHub Container Registry
  - Updates Homebrew formula in `conallob/homebrew-tap`
  - Creates GitHub releases with comprehensive installation instructions

### Release Process

To create a new release:

1. **Tag a new version:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions will automatically:**
   - Run all tests
   - Build binaries for all platforms
   - Create RPM and DEB packages
   - Build and push Docker images to `ghcr.io/conallob/jira-beads-sync`
   - Update Homebrew formula
   - Create a GitHub release with installation instructions

### Required GitHub Secrets

For the release workflow to function properly, configure these secrets in your repository settings:

- **`GITHUB_TOKEN`**: Automatically provided by GitHub Actions (no configuration needed)
  - Used for: Creating releases, publishing to GitHub Container Registry (GHCR)
  - Permissions: `contents: write`, `packages: write` (configured in workflow)

- **`HOMEBREW_TAP_GITHUB_TOKEN`**: Personal Access Token (PAT) for Homebrew tap repository
  - Required: Yes (must be configured manually)
  - Used for: Pushing formula updates to `conallob/homebrew-tap`
  - Required permissions: `repo` (full control of private repositories)
  - Setup:
    1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
    2. Generate new token with `repo` scope
    3. Add as repository secret named `HOMEBREW_TAP_GITHUB_TOKEN`

### Release Artifacts

Each release produces:
- **Binaries**: tar.gz/zip archives for Linux, macOS (Intel/ARM), Windows
- **Linux Packages**:
  - `.deb` files for Debian/Ubuntu (amd64, arm64)
  - `.rpm` files for RHEL/Fedora/CentOS (x86_64, aarch64)
- **Container Images**: Multi-arch Docker images on GHCR
- **Homebrew Formula**: Auto-updated in `conallob/homebrew-tap`
- **Checksums**: `checksums.txt` for verifying downloads

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

By contributing, you agree that your contributions will be licensed under the same license as the project (BSD-3-Clause).
