// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "MrRSS",
    platforms: [
        .macOS(.v14)
    ],
    products: [
        .executable(name: "MrRSS", targets: ["MrRSS"])
    ],
    targets: [
        .executableTarget(
            name: "MrRSS",
            path: "Sources"
        ),
        .testTarget(
            name: "MrRSSTests",
            dependencies: ["MrRSS"],
            path: "Tests"
        )
    ]
)
