# Specification: Automatically switch to ABC input method on activation

## Overview
Neru uses global hotkeys and event taps to intercept keyboard input. When a Vietnamese input method (like JOKey or Telex) is active on macOS, it may compose characters or perform transformations before Neru receives the raw key events. This often leads to unexpected behavior, such as every keypress resulting in the character "a" instead of the intended navigation command.

This track implements an automatic input source switching mechanism. When Neru activates a navigation mode (Grid, Hints, Scroll, etc.), it will automatically switch the macOS input source to "ABC" (standard US English). When the user exits the mode and Neru returns to its idle state, the previous input source will be restored.

## Functional Requirements
- **Portability:** Define a platform-agnostic `InputMethodPort` to manage input source switching, adhering to Neru's Hexagonal Architecture.
- **Current State Detection:** Capture and store the identifier of the currently active macOS input source before switching.
- **Automatic Switching:** Switch to `com.apple.keylayout.ABC` (the standard macOS English layout) immediately upon mode activation.
- **State Restoration:** Restore the previously captured input source identifier when Neru returns to the `Idle` state.
- **Platform Isolation:** Ensure that macOS-specific Carbon TIS (Text Input Source) APIs are encapsulated within a Darwin adapter and do not leak into the core application logic.
- **Graceful Degradation:** On non-macOS platforms, the system should use stub implementations that do not interfere with the application flow.

## Acceptance Criteria
- [ ] When activating any navigation mode (Grid, Hints, Recursive Grid, Scroll), the macOS input method switches to "ABC".
- [ ] When exiting the mode (returning to Idle), the input method restores to whatever was active before activation (e.g., JOKey).
- [ ] The switching occurs *before* the event tap is enabled to ensure the first keypress is correctly handled.
- [ ] The core logic in `internal/app/modes` remains free of Cgo and macOS-specific headers.
- [ ] The project continues to build and run on Linux/Windows (foundations) without errors.

## Out of Scope
- Support for switching to input methods other than "ABC" on activation (this could be a future configuration option).
- Support for input method switching on Linux or Windows in this initial phase.
