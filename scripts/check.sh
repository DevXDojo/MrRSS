#!/bin/bash
# scripts/check.sh - Run all code quality checks

set -e

echo "🔍 Running code quality checks..."

# Backend checks
echo "📦 Checking Go code..."
go vet ./...
echo "✅ Go vet passed"

echo "🎨 Checking Go formatting..."
if [ "$(gofmt -d . | wc -l)" -gt 0 ]; then
    echo "❌ Go code is not formatted. Run 'make format-backend' to fix."
    gofmt -d .
    exit 1
fi
echo "✅ Go formatting OK"

echo "📦 Checking Go imports..."
if [ "$(goimports -d . | wc -l)" -gt 0 ]; then
    echo "❌ Go imports are not formatted. Run 'make format-backend' to fix."
    goimports -d .
    exit 1
fi
echo "✅ Go imports OK"

# macOS client checks
echo "🎨 Checking the macOS client..."
swift build --package-path frontend
echo "✅ macOS client builds"

swift test --package-path frontend
echo "✅ macOS client tests passed"

# Build check
echo "🔨 Checking build..."
go build -v ./...
echo "✅ Build successful"

echo "🎉 All checks passed!"
