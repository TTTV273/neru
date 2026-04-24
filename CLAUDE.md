# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
just build                    # dev build → bin/neru (CGO_ENABLED=1 on macOS)
just release                  # optimized + stripped
just bundle                   # macOS .app bundle
./bin/neru launch             # start background daemon
```

Cross-platform contributor smoke test:
```bash
just build-linux              # CGO_ENABLED=0, pure-Go stubs
just build-windows
```

## Testing

```bash
just test-unit                # pure unit tests (no build tag)
just test-integration         # requires //go:build integration tags
just test                     # unit + integration
just test-race                # both with -race
just test-foundation          # fast cross-platform-safe slice (config, domain, ports)

# Target a single package
go test ./internal/core/domain/hint/
go test -v ./internal/app/services/
go test -tags=integration ./internal/core/infra/accessibility/
```

Test file naming convention:
- `*_test.go` — unit tests (mocks, pure logic, no build tag)
- `*_integration_darwin_test.go` — `//go:build integration && darwin`
- `*_integration_linux_test.go` — `//go:build integration && linux`

## Lint & Format

```bash
just lint     # golangci-lint (also enforces depguard cross-platform rule)
just fmt      # golangci-lint fmt + fix + clang-format for .h/.m files
just vet      # go vet
```

Pre-commit checklist: `just fmt && just lint && just test && just build`

## Architecture

Neru is a background daemon (keyboard → hints/grid/scroll) built on **Hexagonal Architecture**:

```
cmd/neru/               → entry point; main_darwin.go calls runtime.LockOSThread()
internal/cli/           → Cobra commands; IPC client talking to daemon via Unix socket
internal/app/           → application layer: orchestration, modes, services, components
  modes/                → Mode interface implementations (hints, grid, scroll, recursive-grid)
  services/             → business logic orchestration (HintService, GridService, ActionService)
  components/           → per-mode UI components (overlays, indicators)
internal/core/
  domain/               → pure Go: Grid, Hint, Element, Action (zero OS dependencies)
  ports/                → interfaces: AccessibilityPort, OverlayPort, ConfigPort, etc.
  infra/                → adapter implementations
    platform/           → platform factory + darwin/ linux/ windows/ subdirs
    accessibility/      → AXUIElement adapter + cache
    eventtap/           → global keyboard event dispatch
    hotkeys/            → hotkey registration
    ipc/                → Unix socket server/client
    overlay/            → overlay manager
internal/config/        → TOML parsing, validation, defaults
internal/ui/            → coordinate conversion, abstract overlay rendering
```

### The One Rule (enforced by depguard)

> **Non-darwin-tagged code must never import `internal/core/infra/platform/darwin`.**

All platform-specific code lives behind `//go:build darwin` (or `linux`/`windows`). Stubs on unsupported platforms must return `derrors.New(derrors.CodeNotSupported, "...")`.

### Coordinate System

All shared Go code uses **top-left (0,0), Y increases downward**. The darwin adapter (`accessibility_screen_darwin.m`) inverts Cocoa's bottom-left Y before passing data to Go.

### Key Event Flow

1. `eventtap_darwin.m` (Objective-C CGEventTap) captures raw key events
2. `internal/core/infra/eventtap/adapter.go` dispatches to Go
3. `internal/app/modes/handler.go` routes to the active `Mode`
4. Mode calls into a `Service` → `Port` → `Adapter` → macOS API

### Mode Interface

Every navigation mode implements:
```go
type Mode interface {
    Activate(action *string)   // called with h.mu held — use *Locked helpers
    HandleKey(key string)      // called with h.mu held — use exitModeLocked not SetModeIdle
    HandleActionKey(key string)
    Exit()
    ToggleActionMode()
    ModeType() domain.Mode
}
```

Register new modes in `handler.go`'s `modes` map. See `HintsMode` / `GridMode` as reference implementations.

### Adding a New Navigation Mode

1. `internal/core/domain/` — add domain entity/constants
2. `internal/app/services/` — add service
3. `internal/core/infra/` — add infra adapter if needed
4. `internal/app/components/` — add UI component
5. `internal/app/modes/` — implement `Mode` interface; register in `NewHandler()`
6. `internal/cli/` — add CLI command; register in `root.go`
7. `internal/app/ipc_controller.go` — add IPC handler
8. `configs/` + `docs/CONFIGURATION.md` — add hotkey defaults and docs

### Error Handling

Use `internal/core/errors` (package alias `derrors`):
```go
return derrors.New(derrors.CodeNotSupported, "feature X not yet implemented on linux")
// Caller:
if derrors.IsNotSupported(err) { /* log warning, degrade gracefully */ }
```

## Dependency Stack

- Go 1.26+, CGO on macOS (required for Objective-C bridge)
- Cobra (CLI), TOML via BurntSushi/toml, zap (logging), testify (tests)
- `just` (build runner), `golangci-lint` (lint + format), `clang-format` (ObjC)

Logs on macOS: `~/Library/Logs/neru/app.log`

## Commit Convention

`<type>: <subject>` — types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`
