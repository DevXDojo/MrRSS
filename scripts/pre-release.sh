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

echo "📦 Checking frontend dependencies..."
cd frontend
npm audit --audit-level=moderate
echo "✅ Frontend dependencies OK"

# Check version consistency
echo "🏷️  Checking version consistency..."
VERSION=$(grep '"version"' package.json | sed 's/.*"version": "\([^"]*\)".*/\1/')
echo "Frontend version: $VERSION"

cd ..
GO_VERSION=$(grep "const Version" internal/version/version.go | sed 's/.*= "\([^"]*\)".*/\1/')
echo "Backend version: $GO_VERSION"

if [ "$VERSION" != "$GO_VERSION" ]; then
    echo "❌ Version mismatch! Frontend: $VERSION, Backend: $GO_VERSION"
    exit 1
fi

PACKAGE_VERSION=$(grep '"version"' frontend/package.json | sed 's/.*"version": "\([^"]*\)".*/\1/')
echo "Wails version: $PACKAGE_VERSION"

if [ "$VERSION" != "$PACKAGE_VERSION" ]; then
    echo "❌ Version mismatch! Frontend: $VERSION, Wails: $PACKAGE_VERSION"
    exit 1
fi

echo "✅ Version consistency OK"

echo "🎉 Ready for release!"
