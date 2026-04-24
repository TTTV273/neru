# Initial Concept
A keyboard-driven screen navigation tool for macOS (with Linux and Windows support in progress) that allows users to click, scroll, and drag without touching the mouse.

# Product Definition: Neru (練る)

## Vision
Neru (練る — "to refine through practice") aims to be the ultimate keyboard-driven screen navigation tool, allowing users to interact with every pixel and UI element without touching a mouse. It empowers power users, Vim enthusiasts, and individuals with accessibility needs to navigate their OS with extreme speed, precision, and comfort by staying focused on the home row.

## Target Audience
- **Power Users & Vim Enthusiasts:** Developers and keyboard-centric users who want to maximize their efficiency by minimizing hand movement.
- **Accessibility Users:** Individuals with motor impairments or RSI for whom mouse usage is difficult or painful.
- **Open Source Advocates:** Users who value free, transparent, and community-driven software over proprietary alternatives.

## Core Value Propositions
- **Efficiency & Speed:** Navigate, click, and drag with minimal keystrokes and no need to move hands to a mouse/trackpad.
- **Universal Compatibility:** Works in all apps (native, Electron, creative tools) regardless of their accessibility support, thanks to coordinate-based grid systems.
- **Customizability:** Fully configurable via TOML, allowing users to tailor every hotkey and behavior to their specific workflow.

## Key Features & Modes
- **Recursive Grid (Primary):** A high-precision mode that narrows down the screen area recursively for pixel-perfect navigation.
- **Hint Mode:** Smart labels on clickable UI elements for direct interaction in accessibility-enabled applications.
- **Coordinate Grid:** A fast, coarse navigation system using row and column labels.
- **Scroll Mode:** Vim-style scrolling (j/k, d/u) integrated into the navigation flow.
- **Scripting & Automation:** An IPC-based CLI that allows Neru to be integrated into complex automation workflows and hotkey managers.

## Platform Strategy
While the architecture is designed to be cross-platform (Linux/Windows foundations are present), the current priority is **macOS Excellence**. We will focus on providing the most polished, stable, and feature-rich experience on macOS before expanding native implementations to other platforms.
