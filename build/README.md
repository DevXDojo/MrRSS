# Wails v3 Build System Guide

This document describes how to build and package MrRSS using the Wails v3 build system with Task runner.

## Prerequisites

### Required Tools

- **Go 1.25+**: [https://golang.org/dl/](https://golang.org/dl/)
- **Node.js 18+**: [https://nodejs.org/](https://nodejs.org/)
- **Wails CLI v3**: `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha2.117`
- **Task**: [https://taskfile.dev/installation/](https://taskfile.dev/installation/)

### Platform-Specific Dependencies

#### Windows

```powershell
choco install mingw nsis -y
```

#### Linux (Ubuntu/Debian)

```bash
sudo apt-get install -y \
  libgtk-4-dev \
  libwebkitgtk-6.0-dev \
  libsoup-3.0-dev \
  gcc \
  pkg-config
```

#### macOS

```bash
xcode-select --install
```

## Quick Start

### Development Mode

Run the application in development mode with hot reload:

```bash
# Using wails3 dev command
wails3 dev

# Or using task
task dev
```

The dev server will start on port 5173 (configurable via WAILS_VITE_PORT).

### Building for Your Platform

Build for your current platform:

```bash
# Using task (recommended)
task build

# Or directly via platform-specific task
task windows:build  # on Windows
task linux:build    # on Linux
task darwin:build   # on macOS
```

Build output: `build/bin/`

### Packaging

Create installers and packages:

```bash
# Using task
task package

# Platform-specific
task windows:package  # Creates NSIS installer
task linux:package    # Creates AppImage + tar.gz
task darwin:package   # Creates DMG
```

## Advanced Usage

### Cross-Platform Builds

#### Using Docker (Linux/Windows only)

First, build the Docker image (one-time setup):

```bash
task setup:docker
```

Then build for other platforms:

```bash
# From any OS, build for Windows
task windows:build CGO_ENABLED=1

# From any OS, build for Linux
task linux:build CGO_ENABLED=1
```

**Note**: macOS builds should be done on native macOS runners due to signing requirements.

### Architecture-Specific Builds

```bash
# Build for specific architecture
task windows:build ARCH=amd64
task windows:build ARCH=arm64

task linux:build ARCH=amd64
task linux:build ARCH=arm64

task darwin:build ARCH=universal  # Intel + Apple Silicon
```

### Mobile Platforms (Wails v3 only)

**Note**: iOS and Android support is experimental in Wails v3 alpha.

```bash
# iOS (requires macOS + Xcode)
task ios:build

# Android (requires Android SDK)
task android:build
```

## Task Commands Reference

### Common Tasks

- `task build` - Build application for current platform
- `task package` - Package application with installer
- `task run` - Run the built application
- `task dev` - Run in development mode
- `task setup:docker` - Build Docker image for cross-compilation

### Platform-Specific Tasks

#### Windows

- `task windows:build` - Build Windows executable
- `task windows:package` - Create NSIS installer
- `task windows:sign` - Sign executable (requires certificate)

#### Linux

- `task linux:build` - Build Linux binary
- `task linux:package` - Create AppImage + packages
- `task linux:create:appimage` - Create AppImage only
- `task linux:create:tarball` - Create tar.gz only

#### macOS

- `task darwin:build` - Build macOS app bundle
- `task darwin:package` - Create DMG installer
- `task darwin:sign` - Sign app bundle (requires certificate)
- `task darwin:notarize` - Notarize with Apple (requires profile)

### Frontend Tasks

- `task common:install:frontend:deps` - Install frontend dependencies
- `task common:build:frontend` - Build frontend for production
- `task common:dev:frontend` - Run frontend dev server

### Utility Tasks

- `task common:generate:icons` - Generate platform icons from appicon.png
- `task common:generate:bindings` - Generate TypeScript bindings
- `task common:update:build-assets` - Update build assets from config

## Configuration

### Main Configuration

Edit `build/config.yml` to change:

- Application name and version
- Company information
- Bundle identifiers
- Dev mode settings

### Platform-Specific Configuration

Each platform has its own Taskfile in `build/<platform>/Taskfile.yml`:

- `build/windows/Taskfile.yml` - Windows build configuration
- `build/linux/Taskfile.yml` - Linux build configuration
- `build/darwin/Taskfile.yml` - macOS build configuration

### Signing Configuration

#### Windows

Edit `build/windows/Taskfile.yml`:

```yaml
vars:
  SIGN_CERTIFICATE: "path/to/certificate.pfx"
  TIMESTAMP_SERVER: "http://timestamp.digicert.com"
```

Then setup password:

```bash
wails3 setup signing
```

#### macOS

Edit `build/darwin/Taskfile.yml`:

```yaml
vars:
  SIGN_IDENTITY: "Developer ID Application: Your Name (TEAM_ID)"
  NOTARIZATION_PROFILE: "notarization-profile-name"
```

#### Linux

Edit `build/linux/Taskfile.yml`:

```yaml
vars:
  PGP_KEY: "path/to/signing-key.asc"
```

## GitHub Actions

The project includes automated workflows:

### Release Workflow

Triggered manually via GitHub Actions UI:

1. Go to Actions → Release
2. Click "Run workflow"
3. Enter version (e.g., `v1.2.21`)
4. Click "Run workflow"

Builds for all platforms:

- Windows (AMD64, ARM64)
- Linux (AMD64, ARM64)
- macOS (Universal)

### Test Build Workflow

Triggered on push/PR to main:

- Tests build on all platforms
- Validates configuration

## Troubleshooting

### CGO is disabled

**Solution**: CGO must be enabled for Wails v3:

```bash
export CGO_ENABLED=1
task build
```

### Missing dependencies (Linux)

**Solution**: Install all required libraries:

```bash
sudo apt-get update
sudo apt-get install -y libgtk-4-dev libwebkitgtk-6.0-dev libsoup-3.0-dev gcc pkg-config
```

### Task not found

**Solution**: Install Task runner:

```bash
# macOS
brew install go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

# Windows
choco install go-task
```

### Frontend build fails

**Solution**: Clean and reinstall dependencies:

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

## Resources

- [Wails v3 Documentation](https://v3.wails.io/)
- [Task Documentation](https://taskfile.dev/)
- [Build Configuration Reference](./config.yml)
- [GitHub Actions Workflows](../.github/workflows/)

## Support

For issues:

1. Check [GitHub Issues](https://github.com/DevXDojo/MrRSS/issues)
2. Review [Build Requirements](../docs/BUILD_REQUIREMENTS.md)
3. Check [Wails Discord](https://discord.gg/wails)
