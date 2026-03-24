# Contributing to Curspace

Thank you for your interest in contributing to Curspace.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/curspace.git`
3. Create a branch: `git checkout -b my-feature`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit: `git commit -m "Add my feature"`
7. Push: `git push origin my-feature`
8. Open a Pull Request

## Development

```bash
go build ./...
./scripts/modernize.sh
go test ./...
go vet ./...
```

## Guidelines

- Follow standard Go conventions and formatting (`gofmt`)
- Add tests for new functionality
- Keep commits focused and atomic
- Update documentation if adding new commands or features

## Reporting Issues

Use [GitHub Issues](https://github.com/frknikiz/curspace/issues) to report bugs or request features.
