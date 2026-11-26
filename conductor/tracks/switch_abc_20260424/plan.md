# Implementation Plan: Automatically switch to ABC input method on activation

This plan covers the integration of an automatic input method switching mechanism to prevent conflicts with Vietnamese input methods on macOS.

## Phase 1: Infrastructure & Adapter (Foundation)
*The core interface and macOS implementation have been partially drafted. This phase ensures they are correctly integrated into the infra layer.*

- [ ] Task: Finalize `InputMethodPort` interface
    - [ ] Verify `internal/core/ports/input_method.go` exists and has `GetCurrentSourceID()` and `SwitchToSourceID(id string)` methods.
- [ ] Task: Finalize macOS Adapter (Darwin)
    - [ ] Ensure `internal/core/infra/platform/darwin/inputmethod.go` correctly implements the port using Cgo and Carbon APIs.
    - [ ] Verify `inputmethod.h` and `inputmethod_darwin.m` are present and correctly linked.
- [ ] Task: Create Stub Adapters for Other Platforms
    - [ ] Implement no-op `InputMethodAdapter` in `internal/core/infra/platform/linux/` and `internal/core/infra/platform/windows/`.
- [ ] Task: Update Platform Factory
    - [ ] Update the platform infrastructure factory to provide the `InputMethodPort` instance.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Infrastructure & Adapter (Foundation)' (Protocol in workflow.md)

## Phase 2: Application Integration
*Integrating the input method manager into the mode handler and lifecycle.*

- [ ] Task: Update Mode Handler State
    - [ ] Add `InputMethodPort` and `previousInputSource` field to the `Handler` struct in `internal/app/modes/handler.go`.
- [ ] Task: Implement Activation Logic
    - [ ] Update `setModeLocked` in `internal/app/modes/mode_setup.go` to capture current source and switch to ABC *before* enabling the mode.
- [ ] Task: Implement Deactivation Logic
    - [ ] Update `SetModeIdle` in `internal/app/modes/mode_setup.go` to restore the previous input source.
- [ ] Task: Wiring & Dependency Injection
    - [ ] Update `internal/app/app_initialization.go` (or relevant initialization file) to inject the `InputMethodPort` into the mode handler.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Application Integration' (Protocol in workflow.md)

## Phase 3: Verification & Cleanup
*Ensuring stability and correct behavior.*

- [ ] Task: Verify Cross-Platform Compilation
    - [ ] Run `just build` on macOS.
    - [ ] (If possible) Run `GOOS=linux go build ./...` to ensure stub implementations work.
- [ ] Task: Manual Verification on macOS
    - [ ] Launch Neru with a Vietnamese input method active.
    - [ ] Activate Hint mode and verify input method switches to ABC.
    - [ ] Deactivate Hint mode and verify input method restores to Vietnamese.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Verification & Cleanup' (Protocol in workflow.md)
