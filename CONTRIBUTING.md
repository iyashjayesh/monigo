# Contributing to MoniGo

Thank you for your interest in contributing to MoniGo. This document explains the process and standards for contributions.

## Getting Started

```bash
git clone https://github.com/iyashjayesh/monigo.git
cd monigo
go test ./... -race
```

## Development Workflow

1. Fork the repository and create a feature branch from `main`
2. Write your code and tests
3. Ensure all checks pass: `go test ./... -race -cover && go vet ./...`
4. Submit a pull request with a clear description

## Code Standards

- Follow standard Go conventions (`gofmt`, `go vet`)
- All exported symbols must have GoDoc comments
- Tests are required for new functionality
- Use table-driven tests where appropriate
- Run with `-race` flag before submitting

## Pull Request Process

1. PRs must pass CI (tests, vet, race detector)
2. One approval from a maintainer is required
3. Commit messages should be descriptive (not "fix bug" - explain what and why)
4. Breaking changes must be documented in the PR description

## Testing

```bash
# Run all tests with race detector
go test ./... -race -count=1

# Run benchmarks
go test ./... -bench=. -benchmem

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include Go version, OS, and a minimal reproduction case
- For security vulnerabilities, see [SECURITY.md](SECURITY.md)

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
