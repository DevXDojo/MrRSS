# MrRSS SwiftUI Frontend

This directory contains the native macOS 14 SwiftUI frontend for MrRSS. It uses the existing Go HTTP API and does not modify the Vue/Wails frontend.

## Run directly

From the repository root:

```bash
./frontend-swift/run.sh
```

The launcher builds and starts the Go server at `http://127.0.0.1:1234`, waits for the API to become available, and then starts the SwiftUI application. If a server is already listening on that address, the launcher reuses it.

Requirements:

- macOS 14 or later
- Xcode 15 or later
- Go 1.24 or later
- The existing `frontend/dist` directory, which is embedded by the server build

## Run components separately

Start the backend from the repository root:

```bash
go run -tags server . -host 127.0.0.1 -port 1234
```

In another terminal, start the macOS frontend:

```bash
swift run --package-path frontend-swift MrRSS
```

The backend address can be changed in MrRSS Settings. It can also be set before launch:

```bash
MRRSS_API_BASE_URL=http://127.0.0.1:8080/api swift run --package-path frontend-swift MrRSS
```

## Build and test

```bash
swift build --package-path frontend-swift
swift test --package-path frontend-swift
```

To create the universal `.app` bundle and DMG used by GitHub Releases:

```bash
cd frontend
npm ci
npm run build
cd ../frontend-swift
./build-app.sh 1.2.3
```

The release application contains both the SwiftUI frontend and reusable Go backend, so it can launch without a separately installed server.

The application provides native three-column navigation, feed subscription management, folders for grouping subscriptions, source refresh, article filtering, pagination, read/unread and favorite actions, translation, summaries, automation rules, complete backend settings management, configurable server connectivity, and restricted HTML rendering for untrusted feed content.

Folders are the `category` recorded on each feed, so they are shared with the Vue frontend. A subscription moves into a folder by dragging its row onto the folder, or from the row's context menu. A folder that has been created but holds no feeds yet has nowhere to live on the server, so its name is remembered on the Mac until a feed moves into it.
