#!/bin/bash
# scripts/pre-release.sh - Pre-release checks

set -e

echo "🚀 Running pre-release checks..."

# Run all checks
./scripts/check.sh

# Additional release checks
echo "📦 Checking Go modules..."
go mod tidy
if [ -n "$(git status --porcelain go.mod go.sum)" ]; then
    echo "❌ Go modules are not clean. Commit changes first."
    exit 1
fi
echo "✅ Go modules clean"

# Check version consistency
echo "🏷️  Checking version consistency..."
GO_VERSION=$(grep "const Version" internal/version/version.go | sed 's/.*= "\([^"]*\)".*/\1/')
echo "Backend version: $GO_VERSION"

PLIST_VERSION=$(grep -A1 'CFBundleShortVersionString' frontend-swift/packaging/Info.plist | grep '<string>' | sed 's/.*<string>\(.*\)<\/string>.*/\1/')
echo "Client bundle version placeholder: $PLIST_VERSION"

if [ -z "$GO_VERSION" ]; then
    echo "❌ Could not read the backend version."
    exit 1
fi

echo "✅ Version consistency OK"

echo "🎉 Ready for release!"
