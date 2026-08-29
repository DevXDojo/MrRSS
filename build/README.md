# Build Assets

This directory holds the icon assets used when packaging the macOS client.

| File | Purpose |
| --- | --- |
| `appicon.png` | Source icon artwork |
| `darwin/icons.icns` | Icon bundled into `MrRSS-SwiftUI.app` |

## Building the application

The macOS client is a Swift package in `frontend-swift`, and the Go backend is
built as a plain binary that the bundle launches on demand.

```bash
# Build the .app bundle and the DMG
make build-app VERSION=1.3.28

# Or call the script directly
./frontend-swift/build-app.sh 1.3.28
```

The script builds a universal SwiftUI executable, builds the Go backend for
`arm64` and `x86_64`, merges them with `lipo`, copies the icon and `Info.plist`,
signs the bundle, and produces `frontend-swift/dist/MrRSS-<version>-macos-swiftui-universal.dmg`.

Set `MRRSS_BUILD_ARCHS` to limit the architectures, for example
`MRRSS_BUILD_ARCHS=arm64 ./frontend-swift/build-app.sh dev` for a faster local build.

## Requirements

- macOS 14 or later
- Xcode 15 or later
- Go 1.27 or later

See [docs/BUILD_REQUIREMENTS.md](../docs/BUILD_REQUIREMENTS.md) for details.
