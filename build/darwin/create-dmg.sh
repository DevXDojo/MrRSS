#!/bin/bash
# Script to create a macOS DMG installer for MrRSS
#
# Application Information:
# Name: MrRSS
# Description: A Modern, Cross-Platform Desktop RSS Reader
# Publisher: Ch3nyang
# URL: https://github.com/WCY-dt/MrRSS
# Copyright: Copyright © Ch3nyang

set -e

APP_NAME="MrRSS"
# Get version from frontend/package.json if available, otherwise use default
VERSION=$(grep -o '"version"[[:space:]]*:[[:space:]]*"[^"]*"' frontend/package.json 2>/dev/null | head -1 | sed 's/.*"\([^"]*\)".*/\1/' || echo "1.3.19")
APP_PUBLISHER="Ch3nyang"
APP_URL="https://github.com/WCY-dt/MrRSS"
APP_DESCRIPTION="A Modern, Cross-Platform Desktop RSS Reader"
BUILD_DIR="build/bin"
DMG_DIR="build/dmg"
APP_PATH="${BUILD_DIR}/${APP_NAME}.app"
DMG_NAME="${APP_NAME}-${VERSION}-darwin-universal.dmg"

echo "Creating DMG for ${APP_NAME} ${VERSION}..."
echo "Publisher: ${APP_PUBLISHER}"
echo "Description: ${APP_DESCRIPTION}"
echo ""

# Check if app exists
if [ ! -d "${APP_PATH}" ]; then
    echo "Error: Application not found at ${APP_PATH}"
    echo "Please build the application first with: wails3 build -platform darwin/universal"
    exit 1
fi

# Create DMG directory
rm -rf "${DMG_DIR}"
mkdir -p "${DMG_DIR}"

# Copy app to DMG directory
echo "Copying application..."
cp -R "${APP_PATH}" "${DMG_DIR}/"

# Create Applications symlink
echo "Creating Applications symlink..."
ln -s /Applications "${DMG_DIR}/Applications"

# Create DMG
echo "Creating DMG image..."
rm -f "${BUILD_DIR}/${DMG_NAME}"

# Use hdiutil to create the DMG
if ! hdiutil create -volname "${APP_NAME}" \
    -srcfolder "${DMG_DIR}" \
    -ov -format UDZO \
    "${BUILD_DIR}/${DMG_NAME}"; then
    echo "Error: Failed to create DMG with hdiutil"
    echo "This might be due to permissions or disk space issues"
    exit 1
fi

# Verify the DMG was created
if [ ! -f "${BUILD_DIR}/${DMG_NAME}" ]; then
    echo "Error: DMG file was not created at ${BUILD_DIR}/${DMG_NAME}"
    exit 1
fi

# Clean up
rm -rf "${DMG_DIR}"

echo "DMG created successfully: ${BUILD_DIR}/${DMG_NAME}"
echo ""
echo "Installation instructions:"
echo "1. Open the DMG file"
echo "2. Drag ${APP_NAME}.app to the Applications folder"
echo "3. Launch ${APP_NAME} from Applications"
echo ""
echo "User data will be stored in: ~/Library/Application Support/MrRSS/"
