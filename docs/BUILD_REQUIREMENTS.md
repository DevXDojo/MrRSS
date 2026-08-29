# Build Requirements

This branch builds the macOS client and the Go backend behind it. There is no
web frontend and no desktop shell to compile.

## Requirements

- **macOS 14** or later
- **Xcode 15** or later, for the Swift toolchain
- **Go 1.27** or later

Check what you have:

```bash
sw_vers -productVersion
swift --version
go version
```

If `swift` is missing, install the command line tools:

```bash
xcode-select --install
```

## CGO

The backend uses `modernc.org/sqlite`, a pure Go implementation, so CGO is not
required. The packaging script builds the backend with `CGO_ENABLED=0` for both
architectures and merges them with `lipo`.

Running the backend tests does use CGO in continuous integration, which is why
the workflow sets `CGO_ENABLED=1` there.

## Building

### The client alone

```bash
swift build --package-path frontend
swift test --package-path frontend
```

### The backend alone

```bash
go build -o bin/mrrss-server .
go test ./internal/...
```

### Both, and the application bundle

```bash
make build                        # backend binary and client executable
make build-app VERSION=1.3.28     # signed .app bundle and universal DMG
```

The packaging script builds a universal SwiftUI executable, builds the backend
for `arm64` and `x86_64`, merges them, copies the icon and `Info.plist`, signs
the bundle and produces
`frontend/dist/MrRSS-<version>-macos-swiftui-universal.dmg`.

To build for one architecture while developing:

```bash
MRRSS_BUILD_ARCHS=arm64 ./frontend/build-app.sh dev
```

## Running during development

```bash
./frontend/run.sh
```

The launcher builds the backend, starts it on `http://127.0.0.1:1234`, waits for
the API to answer and then starts the client. An existing server on that address
is reused.

To run the halves separately:

```bash
go run . -host 127.0.0.1 -port 1234
MRRSS_API_BASE_URL=http://127.0.0.1:1234/api swift run --package-path frontend MrRSS
```

## The server on other platforms

The backend itself is portable. Building it for Linux needs nothing beyond Go:

```bash
GOOS=linux GOARCH=amd64 go build -o mrrss-server .
```

Or use the Docker image:

```bash
docker build -f Dockerfile.server -t mrrss-server:latest .
```

## Troubleshooting

**`swift build` cannot find the toolchain**

Point `xcode-select` at a full Xcode installation:

```bash
sudo xcode-select -s /Applications/Xcode.app/Contents/Developer
```

**The build stalls while compiling a view**

SwiftUI bodies are type-checked as a whole, and a long expression can take a very
long time. Split the body into smaller computed properties.

**`codesign` fails during packaging**

The script signs ad hoc, which needs no certificate. If signing still fails,
check that the bundle is not on a volume that strips extended attributes.
