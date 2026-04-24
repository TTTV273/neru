# Tech Stack: Neru (練る)

## Core Technologies
- **Language:** **Go 1.26.1** — Chosen for its performance, concurrency primitives, and ease of cross-compilation.
- **CLI Framework:** **Cobra** — Provides a robust structure for CLI commands, flags, and integrated help.
- **Logging:** **Zap** — A high-performance, structured logging library from Uber, ideal for both development and production logs.
- **Configuration:** **TOML (via BurntSushi/toml)** — Used for user-facing configuration due to its readability and support for complex nested structures.

## Architecture & Integration
- **Architecture:** **Hexagonal (Ports and Adapters)** — Ensures a clean separation between the core navigation logic and platform-specific implementations (macOS, Linux, Windows).
- **OS Interfacing:** **Cgo & Objective-C** — Required for direct interaction with low-level macOS APIs (Carbon, Accessibility, Quartz) to capture events and render overlays.
- **Build System:** **Just** — A command runner used to automate builds, testing, linting, and release processes.

## Development Tools
- **Testing:** **Go's standard `testing` package + `stretchr/testify`** — Used for both unit tests (logic) and integration tests (OS interactions).
- **Linting:** **golangci-lint** — Aggregates multiple linters to ensure code quality and consistency.
