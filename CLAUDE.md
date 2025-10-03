# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Peacock is a CI/CD tool for communicating release notes to users. It parses PR descriptions with "Notify" headers to send targeted messages to different teams via Slack or webhooks. The project supports both CLI and API modes.

## Architecture

The codebase follows clean architecture patterns with these key layers:

### Core Components
- **Domain Layer** (`pkg/domain/`): Interfaces and business logic abstractions
  - `ReleaseNotesUseCase`: Parses markdown, manages release notes, sends notifications
  - `ReleaseUseCase`: Manages release data persistence
  - `FeathersUseCase`: Handles team configuration from `.peacock/feathers.yaml`
  - `Git` and `SCM` interfaces: Git operations and GitHub API interactions

### Use Cases (`pkg/*/usecase/`)
- **Release Notes UC** (`pkg/releasenotes/usecase/`): Core business logic for parsing PR descriptions and sending notifications
- **Release UC** (`pkg/release/usecase/`): Manages release data storage and retrieval
- **Feathers UC** (`pkg/feathers/`): Loads and validates team configurations

### Infrastructure
- **Git Integration** (`pkg/git/`): Local git operations and GitHub API client
- **Message Clients** (`pkg/msgclients/`): Slack and webhook notification handlers
- **Server** (`pkg/server/`): HTTP API server with dependency injection
- **MongoDB Repositories** (`pkg/*/repository/mongodb/`): Data persistence layer

### Entry Points
- **CLI** (`cmd/main.go` → `cmd/cli/main.go`): Command-line interface
- **API Server** (`cmd/api/main.go`): HTTP server for webhook endpoints

## Configuration

- **Feathers**: Team configurations stored in `.peacock/feathers.yaml` in target repositories
- **Environment Variables**: See README.md for complete list (GITHUB_TOKEN, SLACK_TOKEN, etc.)
- **Config Package** (`pkg/config/`): Centralized configuration management

## Common Development Commands

### Building
```bash
make build          # Build both CLI and API binaries
make build-cli      # Build CLI binary only
make install        # Install CLI binary to $GOPATH/bin
```

### Testing
```bash
make test           # Run unit tests
make test-coverage  # Run tests with coverage
make test-report    # Generate coverage report
```

### Code Quality
```bash
make fmt            # Format code and imports
make lint           # Run linters (calls ./hack/gofmt.sh, ./hack/linter.sh, ./hack/generate.sh)
make all            # Format, build, test, and lint
```

### Development Utilities
```bash
make mocks          # Generate mock implementations from domain interfaces
make swag           # Generate Swagger documentation
make docs           # Generate CLI documentation
```

### Cross-Platform Building
```bash
make linux          # Build for Linux
make darwin         # Build for macOS
make win            # Build for Windows
```

## Key Patterns

### Message Processing Flow
1. Parse PR description for `### Notify` headers
2. Extract team names and message content
3. Validate teams against feathers configuration
4. Convert markdown to appropriate format (Slack/HTML)
5. Send via configured message clients

### Dependency Injection
The server uses constructor injection pattern in `pkg/server/inject.go` to wire dependencies.

### Testing
- Unit tests use `--tags=unit` build tag
- Mocks are generated in `pkg/domain/mocks/` using mockery
- Test utilities in `pkg/utils/testUtils.go`

## Important Files

- `pkg/cmd/run/run.go`: Main CLI command implementation
- `pkg/releasenotes/usecase/releasenotesuc.go`: Core business logic
- `pkg/markdown/markdown.go`: Markdown parsing and conversion
- `development-config.yaml`: Local development configuration