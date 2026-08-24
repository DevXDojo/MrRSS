# Build Requirements

This document describes the system-level dependencies required for building MrRSS on different platforms.

## Overview

MrRSS uses Wails v3 (alpha) framework which requires CGO (C bindings for Go):

- **Wails v3**: For the desktop application framework with built-in system tray
- **SQLite**: Pure Go implementation (`modernc.org/sqlite`), no C dependencies

## Important: CGO Requirement

⚠️ **CRITICAL**: Wails v3 requires CGO to be enabled. You must set:

```bash
export CGO_ENABLED=1
```

Or when building:

```bash
CGO_ENABLED=1 wails3 build
```

## Platform-Specific Requirements

### Linux

#### Development Dependencies

```bash
sudo apt-get update
sudo apt-get install -y \
  gcc \
  pkg-config \
  libgtk-4-dev \
  libwebkitgtk-6.0-dev \
  libsoup-3.0-dev
```

**Dependency Breakdown**:

- `gcc`: C compiler (required for CGO)
- `pkg-config`: Build tool for finding libraries
- `libgtk-4-dev`: GTK4 development headers (for Wails UI)
- `libwebkitgtk-6.0-dev`: WebKitGTK 6.0 development headers (for Wails webview, **required for current Wails v3**)
- `libsoup-3.0-dev`: HTTP library 3.0 (required for Wails v3)

**Important**: Current Wails v3 requires GTK4, WebKitGTK 6.0, and libsoup 3.0. Older WebKitGTK 4.x and GTK3 packages are not sufficient.

**Note for Linux Mint**: Also install `libxapp-dev`

#### Runtime Dependencies

End users running the compiled binary will need:

- `libgtk-4-1`
- `libwebkitgtk-6.0-4`
- `libsoup-3.0-0`

### Windows

#### Development Dependencies

Install via Chocolatey:

```powershell
choco install mingw nsis -y
```

**Dependency Breakdown**:

- `mingw`: MinGW-w64 GCC compiler (required for CGO)
- `nsis`: Nullsoft Scriptable Install System (for creating installers)

#### Alternative: Manual Installation

If not using Chocolatey:

1. Install [MinGW-w64](https://www.mingw-w64.org/)
2. Install [NSIS](https://nsis.sourceforge.io/) (optional, for installers)
3. Add MinGW `bin` directory to PATH

#### Build Flags

To avoid opening a console at application startup:

```bash
go build -ldflags "-H=windowsgui"
```

Or with Wails:

```bash
wails3 build -ldflags "-H=windowsgui"
```

#### Runtime Dependencies

Windows binaries are statically linked and don't require additional runtime dependencies.

### macOS

#### Development Dependencies

Install Xcode Command Line Tools (if not already installed):

```bash
xcode-select --install
```

**Note**: macOS has native support for systray through AppKit, so no additional libraries are needed.

#### Application Bundle

macOS requires an application bundle structure:

```plaintext
MrRSS.app/
  Contents/
    Info.plist
    MacOS/
      MrRSS
    Resources/
      MrRSS.icns
```

Wails automatically creates this structure during build.

#### Info.plist Settings

Add these keys for better macOS integration:

```xml
<!-- High resolution support -->
<key>NSHighResolutionCapable</key>
<string>True</string>

<!-- Hide from Dock (optional, for menu bar only apps) -->
<key>LSUIElement</key>
<string>1</string>
```

#### Runtime Dependencies

macOS binaries are self-contained and don't require additional runtime dependencies.

## Building with Wails

### Standard Build

```bash
# Development build with hot reload
wails3 dev

# Production build (recommended: use Task)
task build

# Or directly with wails3
wails3 build

# Platform-specific build with Task
task linux:build
task windows:build
task darwin:build
```

### Build Configuration

Wails v3 uses `build/config.yml` for build configuration and Taskfile for platform-specific builds:

- **Frontend**: Automatically built via `frontend/package.json` scripts
- **Backend**: CGO-enabled Go build with platform-specific flags
- **Installers**: Created via platform-specific scripts (NSIS, create-dmg.sh, create-appimage.sh)

### Cross-Compilation

**Note**: Cross-compilation with CGO is complex. For best results:

- Build Linux binaries on Linux
- Build Windows binaries on Windows
- Build macOS binaries on macOS

GitHub Actions handles this automatically using platform-specific runners.

## GitHub Actions

Our CI/CD pipeline automatically installs all required dependencies:

### Test Workflow

- Installs Linux dependencies for backend tests
- Sets `CGO_ENABLED=1`

### Release Workflow

- Platform-specific dependency installation
- Cross-platform builds using native runners
- Artifact creation (installers, AppImages, DMGs)

## Troubleshooting

### "CGO is disabled" Error

**Solution**: Enable CGO before building:

```bash
export CGO_ENABLED=1
wails3 build
```

### Linux: "Package webkitgtk-6.0 was not found"

**Solution**: Install WebKitGTK 6.0 development headers:

```bash
sudo apt-get install libwebkitgtk-6.0-dev
```

### Linux: "Package ayatana-appindicator3-0.1 was not found"

This error is from older versions. Wails v3 uses its own system tray implementation.

### Linux: "Package libsoup-3.0 was not found"

**Solution**: Install libsoup3 development headers:

```bash
sudo apt-get install libsoup-3.0-dev
```

### Windows: "gcc: command not found"

**Solution**: Install MinGW:

```powershell
choco install mingw -y
```

Or download from [mingw-w64.org](https://www.mingw-w64.org/) and add to PATH.

### macOS: Missing Xcode Command Line Tools

**Solution**: Install Xcode Command Line Tools:

```bash
xcode-select --install
```

## Development Environment Setup

### Quick Setup Scripts

**Linux/macOS**:

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..

# Run development server
wails3 dev
```

**Windows (PowerShell)**:

```powershell
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..

# Run development server
wails3 dev
```

## Related Documentation

- [Architecture Overview](ARCHITECTURE.md)
- [Code Patterns](CODE_PATTERNS.md)
- [Testing Guide](TESTING.md)
