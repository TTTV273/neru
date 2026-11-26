# Product Guidelines: Neru (練る)

## Tone & Voice
Neru's communication—whether in documentation or system messages—should be **empathetic and guiding**. We acknowledge that keyboard-based navigation has a learning curve and may be used by individuals with motor impairments. We aim to be helpful and encouraging, providing clear paths for users to master the tool without being overly verbose.

## User Experience (UX) Principles
- **Keyboard-First Design:** Every feature and configuration must be fully accessible via the keyboard. The mouse is an optional fallback, never a requirement.
- **Visual Minimalism:** Overlays (hints, grids, status indicators) should be as non-intrusive as possible. They should provide necessary feedback without cluttering the screen or obscuring the user's primary workspace.
- **Performance-Centric UX:** Responsiveness is a feature. All interactions must have near-zero latency to ensure that the tool keeps pace with the user's thought process and muscle memory.

## Documentation Standards
- **Markdown-Centric:** All documentation lives in the repository as Markdown files, ensuring it is version-controlled alongside the code and accessible directly in the terminal or code editor.
- **Example-Heavy:** We prioritize practical "how-to" examples and scenarios to show users exactly how to achieve their goals.
- **Deep Technical Deep-Dives:** For contributors and curious power users, we provide exhaustive documentation on Neru's internal mechanics, architecture (Hexagonal/Ports & Adapters), and platform-specific implementations.

## Error Handling & Communication
- **Silent/Unobtrusive Errors:** Neru should favor failing gracefully or silently to avoid disrupting the user's focus. 
- **Log-First Diagnostics:** Detailed error information should be piped to system logs (using `zap`) for power users to diagnose, rather than being presented in disruptive pop-ups.
- **Actionable Guidance:** When a user-facing error is necessary (e.g., missing system permissions), the message should be empathetic and provide a clear, direct step to resolve the issue.
