# Project Overview

**Neru** (練る) is a keyboard-driven screen navigation tool for macOS (with Linux and Windows support in progress) that allows users to click, scroll, and drag without touching the mouse. It serves as a free, open-source alternative to tools like Homerow and Vimac.

The project is built with **Go** (1.26+) and utilizes **Cgo + Objective-C** to interface with native macOS APIs (Accessibility, CGEventTap). 

## Core Architecture

Neru follows a **Hexagonal Architecture (Ports and Adapters)** to cleanly separate business logic from platform-specific implementations:
*   **Domain (`internal/core/domain`)**: Pure Go business logic (Grid, Hints, Actions).
*   **Ports (`internal/core/ports`)**: Interfaces defining OS capabilities (Accessibility, EventTap, Overlays).
*   **Adapters (`internal/core/infra`)**: OS-specific implementations of ports. Non-macOS builds are currently mostly stubs or foundations.
*   **Application (`internal/app`)**: Orchestrates services and navigation modes (Hints, Grid, Recursive Grid, Scroll).
*   **CLI (`internal/cli`)**: Built with **Cobra**, handles user commands and communicates with the background daemon via **Unix Domain Sockets (IPC)**.

## Building and Running

The project uses [Just](https://github.com/casey/just) as its build automation tool.

### Prerequisites
*   Go 1.26+
*   `just` (Command runner)
*   `golangci-lint` (Linter)
*   macOS (for full features; Linux/Windows builds are foundations only)

### Key Commands

*   **Build (Development):** `just build` (compiles to `bin/neru`)
*   **Build (Release):** `just release` (optimized, stripped binary)
*   **Run Daemon:** `./bin/neru launch`
*   **Trigger an Action:** `neru hints` or `neru grid`
*   **Test (Unit):** `just test-unit`
*   **Test (Integration):** `just test-integration` (requires `//go:build integration` tags)
*   **Test All:** `just test`
*   **Format Code:** `just fmt`
*   **Lint Code:** `just lint`

## Development Conventions

*   **Cross-Platform Discipline:** The "One Rule" is that non-darwin tagged code must *never* import `internal/core/infra/platform/darwin`. Platform isolation is strictly enforced via build tags (e.g., `//go:build darwin`).
*   **Coordinate System:** Neru uses a global top-left (0,0) coordinate system. The macOS adapter handles inverting Cocoa's bottom-left Y-axis before passing data to Go.
*   **Graceful Degradation:** Unsupported platform features return a specific `derrors.CodeNotSupported` error instead of failing silently.
*   **Testing:** Tests are clearly separated into `*_test.go` (pure logic/mocks) and `*_integration_darwin_test.go` (real OS APIs).
*   **Commit Messages:** Follows conventional commits (`feat:`, `fix:`, `refactor:`, etc.).
*   **Code Style:** Managed by `golangci-lint`, `goimports`, and `.editorconfig`. Objective-C code uses `clang-format`.

## User Goals
- **Mục tiêu dài hạn:** Trở thành tỷ phú thông qua việc phát triển các sản phẩm và hệ thống AI.
- **Chiến lược hiện tại:** Xây dựng kênh YouTube hướng dẫn lập trình AI Agent và làm các dự án Freelance về AI.
- **Mục tiêu với dự án Neru:** Dùng dự án Neru (viết bằng Go) làm môi trường thực chiến để rèn luyện kỹ năng lập trình Go (vừa mới học). Tìm hiểu kiến trúc Hexagonal và thêm một vài tính năng mới theo ý tưởng cá nhân.
