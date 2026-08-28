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

## Releasing

Maintainers only.

MoniGo is on the **v1 line**. `go.mod` declares
`module github.com/iyashjayesh/monigo` with no major-version suffix, and Go
requires the major version to appear in the module path from v2 onwards. A `v2.x`
tag on this module path is therefore invalid and the proxy will refuse it -- which
is what happened to `v2.0.0`, leaving `go get` serving `v1.2.0` for months while
several releases' worth of work sat unreachable.

So: **tag `v1.x.y`, and verify the proxy actually served it.** The second half is
the step whose absence caused the problem.

```bash
# 1. CHANGELOG.md: move [Unreleased] entries under a dated version heading
# 2. tag and push
git tag -a v1.3.0 -m "v1.3.0"
git push origin v1.3.0

# 3. VERIFY -- this is not optional
curl -s https://proxy.golang.org/github.com/iyashjayesh/monigo/@latest
# must report the version just tagged

# 4. confirm it installs from a clean module
cd "$(mktemp -d)" && go mod init tmp >/dev/null
go get github.com/iyashjayesh/monigo@latest && go list -m github.com/iyashjayesh/monigo
```

If step 3 does not show the new version, the tag is not usable and the release has
not actually happened, however green the CI run was.

Moving to a v2 line later means renaming the module to
`github.com/iyashjayesh/monigo/v2` in `go.mod` and updating every import path in
the examples and README. That is a breaking change for every consumer, so it needs
a reason beyond the version number looking tidy.

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include Go version, OS, and a minimal reproduction case
- For security vulnerabilities, see [SECURITY.md](SECURITY.md)

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
