<h1><img src="imgs/logo-rounded.png" alt="MrRSS logo" style="height: 40px;"/>&nbsp;MrRSS</h1>

<a href="https://trendshift.io/repositories/15731" target="_blank"><img src="https://trendshift.io/api/badge/repositories/15731" alt="DevXDojo%2FMrRSS | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>

![Screenshot](imgs/og1.png)

<p>
   <strong>English</strong> | <a href="README_zh.md">简体中文</a>
</p>

> **This branch builds the macOS client.** The interface is a native SwiftUI
> application in `frontend`, and the Go backend runs as a plain HTTP API
> server behind it. The Vue frontend and the Wails shell are not part of this
> branch.

[![Version](https://img.shields.io/badge/version-1.3.28-blue.svg)](https://github.com/DevXDojo/MrRSS/releases)
[![License](https://img.shields.io/badge/license-GPLv3-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go)](https://go.dev/)
[![Swift](https://img.shields.io/badge/Swift-5.9+-F05138?logo=swift)](https://swift.org/)
[![macOS](https://img.shields.io/badge/macOS-14+-000000?logo=apple)](https://www.apple.com/macos/)

## ✨ Features

- 🌐 **Auto-Translation & Summarization**: Automatically translate article titles and content, and generate concise summaries to help you get information quickly
- 🤖 **AI-Enhanced Features**: Integrated advanced AI technology for translation, summarization, recommendations, and more. Reading and performing operations through skills are also supported
- 🔌 **Rich Plugin Ecosystem**: Supports integration with mainstream tools like Obsidian, Notion, FreshRSS, and RSSHub for easy feature extension
- 📡 **Diverse Subscription Methods**: Supports URL, XPath, scripts, newsletters, and other feed types to meet different needs
- 🏭 **Custom Scripts & Automation**: Built-in filters and scripting system supporting highly customizable automation workflows

## 🚀 Quick Start

### Download and Install

#### Option 1: Download Pre-built Installer (Recommended)

Download the latest installer for your platform from the [Releases](https://github.com/DevXDojo/MrRSS/releases/latest) page.

<details>

<summary>Click to view the list of available installers</summary>

<div markdown="1">

**Standard Installation:**

- **Windows:** `MrRSS-{version}-windows-amd64-installer.exe` / `MrRSS-{version}-windows-arm64-installer.exe`
- **macOS:** `MrRSS-{version}-darwin-universal.dmg`
- **Linux:** `MrRSS-{version}-linux-amd64.AppImage` / `MrRSS-{version}-linux-arm64.AppImage`

**Portable Version** (no installation required, all data in one folder):

- **Windows:** `MrRSS-{version}-windows-{arch}-portable.zip`
- **Linux:** `MrRSS-{version}-linux-{arch}-portable.tar.gz`
- **macOS:** `MrRSS-{version}-darwin-{arch}-portable.zip`

**AI Agent Skills:**

- **Codex:** `MrRSS-{version}-skills.zip` ([usage guide](docs/SKILLS.md))

</div>

</details>

#### Option 2: Build from Source

<details>

<summary>Click to expand the build from source guide</summary>

<div markdown="1">

### Prerequisites

- [Go](https://go.dev/) 1.27 or higher
- macOS 14 or later
- Xcode 15 or later (for the Swift toolchain)

See [Build Requirements](docs/BUILD_REQUIREMENTS.md) for details.

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/DevXDojo/MrRSS.git
   cd MrRSS
   ```

2. **Run the client**

   ```bash
   ./frontend/run.sh
   ```

   The launcher builds the Go backend, starts it on `http://127.0.0.1:1234`,
   waits for the API to answer, and then starts the client. An existing server
   on that address is reused.

3. **Or run the two halves separately**

   ```bash
   go run . -host 127.0.0.1 -port 1234
   swift run --package-path frontend MrRSS
   ```

   The backend address can be changed in Settings, or before launch:

   ```bash
   MRRSS_API_BASE_URL=http://127.0.0.1:8080/api swift run --package-path frontend MrRSS
   ```

4. **Build the application bundle**

   ```bash
   make build-app VERSION=1.3.28
   ```

   This produces `frontend/dist/MrRSS-SwiftUI.app` and a universal DMG
   beside it. The bundle carries the backend, so it launches without a
   separately installed server.

</div>

</details>

### Server Mode

<details>

<summary>Click to expand the server guide</summary>

<div markdown="1">

The Go binary is an HTTP API server. Run it on its own to share one library
between machines, and point the client at it in Settings:

```bash
go build -o mrrss-server .
./mrrss-server -host 0.0.0.0 -port 1234
```

A Docker image is available too:

```bash
docker build -f Dockerfile.server -t mrrss-server:latest .
docker run -p 1234:1234 -v $PWD/data:/app/data mrrss-server:latest
```

The API is documented in [docs/SERVER_MODE/swagger.json](docs/SERVER_MODE/swagger.json).

</div>

</details>

### Data Storage

<details>

<summary>Click to expand data storage details</summary>

<div markdown="1">

The backend keeps its database, logs and scripts in a `data` directory. Which
one it uses depends on how it was started, and **the copies are independent of
each other**:

| How it was started | Data directory |
| --- | --- |
| `./frontend/run.sh` | `data/` in the repository |
| The packaged `.app` | `~/Library/Application Support/MrRSS-SwiftUI/data/` |
| `go run .` or the built binary | `data/` beside the directory it was started from |
| The Docker image | `/app/data` in the container |

So the library you build up while running from source is not the one the
installed application reads. To carry one across, quit both and copy
`data/rss.db` (along with `rss.db-shm` and `rss.db-wal` if they are present).

Because the application's data lives outside the bundle, removing the
application leaves the library in place; delete the directory above to remove it
as well.

</div>

</details>

## 🛠️ Development Guide

<details>

<summary>Click to expand the development guide</summary>

<div markdown="1">

### Running in Development Mode

```bash
# Backend and client together
make dev

# Backend only, with debug logging
make serve

# Client only, against a running backend
swift run --package-path frontend MrRSS
```

### Code Quality Tools

#### Using Make

A `Makefile` covers the common development tasks:

```bash
# Show all available commands
make help

# Run full check (lint + test + build)
make check

# Clean build artifacts
make clean

# Setup development environment
make setup
```

### Pre-commit Hooks

This project uses pre-commit hooks to ensure code quality:

```bash
# Install hooks
pre-commit install

# Run on all files
pre-commit run --all-files
```

### Running Tests

```bash
make test
```

### Local API

The bundled backend serves its REST API at `http://localhost:1234/api`, and the
listener is restricted to the local computer. Point the client at another
address in Settings to read a shared library.

```bash
# Using Docker
docker run -p 1234:1234 mrrss-server:latest

# Or build from source
go build -o mrrss-server .
./mrrss-server
```

Pre-built server images based on ghcr.io are also available:

```bash
docker run -d -p 1234:1234 ghcr.io/devxdojo/mrrss:latest-amd64
docker run -d -p 1234:1234 ghcr.io/devxdojo/mrrss:latest-arm64
```

Please refer to the [Server Mode API Documentation](docs/SERVER_MODE/swagger.json) for a complete API reference.
To let Codex operate MrRSS through this API, install the release skill package described in [MrRSS Skills](docs/SKILLS.md).

</div>

</details>

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

<details>

<summary>Click to expand the contributing guidelines</summary>

<div markdown="1">

Before contributing:

1. Read the [Code of Conduct](CODE_OF_CONDUCT.md)
2. Check existing issues or create a new one
3. Fork the repository and create a feature branch
4. Make your changes and add tests
5. Submit a pull request

</div>

</details>

## 🔒 Security

If you discover a security vulnerability, please follow our [Security Policy](SECURITY.md).

## 📝 License

This project is licensed under the GPL-3.0 License - see the [LICENSE](LICENSE) file for details.

## 📮 Contact & Support

- **Issues**: [GitHub Issues](https://github.com/DevXDojo/MrRSS/issues)
- **Discussions**: [GitHub Discussions](https://github.com/DevXDojo/MrRSS/discussions)
- **Repository**: [github.com/DevXDojo/MrRSS](https://github.com/DevXDojo/MrRSS)

<div align="center">
  <img src="imgs/sponsor.png" alt="Sponsor MrRSS"/>
  <p>Made with ❤️ by the MrRSS Team</p>
  <p>⭐ Star us on GitHub if you find this project useful!</p>
</div>
