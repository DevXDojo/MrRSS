#!/bin/bash
# Script to create a Linux AppImage for MrRSS
#
# Application Information:
# Name: MrRSS
# Description: A Modern, Cross-Platform Desktop RSS Reader
# Publisher: Ch3nyang
# URL: https://github.com/WCY-dt/MrRSS
# Copyright: Copyright © Ch3nyang

# Exit on error, but allow some commands to fail gracefully
set -e

APP_NAME="MrRSS"
# Get version from frontend/package.json if available, otherwise use default
VERSION=$(grep -o '"version"[[:space:]]*:[[:space:]]*"[^"]*"' frontend/package.json 2>/dev/null | head -1 | sed 's/.*"\([^"]*\)".*/\1/' || echo "1.3.19")
# Get architecture from environment variable or default to amd64
ARCH=${ARCH:-amd64}
echo "Target architecture: ${ARCH}"
echo "System architecture: $(uname -m)"
APP_PUBLISHER="Ch3nyang"
APP_URL="https://github.com/WCY-dt/MrRSS"
APP_DESCRIPTION="A Modern, Cross-Platform Desktop RSS Reader"
BUILD_DIR="build/bin"
APPDIR="build/appimage/${APP_NAME}.AppDir"
APPIMAGE_NAME="${APP_NAME}-${VERSION}-linux-${ARCH}.AppImage"

echo "Creating AppImage for ${APP_NAME} ${VERSION}..."
echo "Publisher: ${APP_PUBLISHER}"
echo "Description: ${APP_DESCRIPTION}"
echo ""

# Check if binary exists
if [ ! -f "${BUILD_DIR}/${APP_NAME}" ]; then
    echo "Error: Binary not found at ${BUILD_DIR}/${APP_NAME}"
    echo "Please build the application first with: wails3 build -platform linux/amd64"
    exit 1
fi

# Create AppDir structure
echo "Creating AppDir structure..."
rm -rf "build/appimage"
mkdir -p "${APPDIR}/usr/bin"
mkdir -p "${APPDIR}/usr/share/applications"
mkdir -p "${APPDIR}/usr/share/icons/hicolor/256x256/apps"

# Copy binary
echo "Copying binary..."
cp "${BUILD_DIR}/${APP_NAME}" "${APPDIR}/usr/bin/"
chmod +x "${APPDIR}/usr/bin/${APP_NAME}"

# Create desktop file
echo "Creating desktop file..."
cat > "${APPDIR}/usr/share/applications/${APP_NAME}.desktop" << EOF
[Desktop Entry]
Type=Application
Name=${APP_NAME}
GenericName=RSS Reader
Comment=${APP_DESCRIPTION}
Exec=${APP_NAME}
Icon=${APP_NAME}
Categories=Network;News;Feed;
Terminal=false
StartupWMClass=${APP_NAME}
Keywords=RSS;Atom;Feed;News;Reader;
X-GNOME-UsesNotifications=true
EOF

# Create AppRun script
echo "Creating AppRun script..."
cat > "${APPDIR}/AppRun" << 'EOF'
#!/bin/bash
SELF=$(readlink -f "$0")
HERE=${SELF%/*}
export PATH="${HERE}/usr/bin:${PATH}"
export LD_LIBRARY_PATH="${HERE}/usr/lib:${LD_LIBRARY_PATH}"
exec "${HERE}/usr/bin/MrRSS" "$@"
EOF
chmod +x "${APPDIR}/AppRun"

# Copy icon (if exists, otherwise create placeholder)
# Icon handling is non-critical - continue even if it fails
set +e
if [ -f "imgs/logo.svg" ] && (command -v inkscape &> /dev/null || command -v convert &> /dev/null); then
    echo "Converting icon..."
    # If inkscape is available, convert SVG to PNG
    if command -v inkscape &> /dev/null; then
        inkscape "imgs/logo.svg" -o "${APPDIR}/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" -w 256 -h 256 2>/dev/null || echo "Warning: inkscape icon conversion failed"
        cp "${APPDIR}/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" "${APPDIR}/${APP_NAME}.png" 2>/dev/null || true
    elif command -v convert &> /dev/null; then
        convert -background none -size 256x256 "imgs/logo.svg" "${APPDIR}/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" 2>/dev/null || echo "Warning: ImageMagick icon conversion failed"
        cp "${APPDIR}/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" "${APPDIR}/${APP_NAME}.png" 2>/dev/null || true
    fi
elif [ -f "build/appicon.png" ]; then
    # Fallback to pre-built PNG icon from Wails build process
    echo "Using existing PNG icon..."
    cp "build/appicon.png" "${APPDIR}/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" 2>/dev/null || echo "Warning: Failed to copy icon"
    cp "build/appicon.png" "${APPDIR}/${APP_NAME}.png" 2>/dev/null || true
else
    echo "Warning: No icon available - AppImage will be created without an icon"
fi
set -e

# Copy desktop file to root
cp "${APPDIR}/usr/share/applications/${APP_NAME}.desktop" "${APPDIR}/"

# Set APPIMAGE_ARCH early for checks
APPIMAGE_ARCH="${ARCH}"
if [ "${ARCH}" = "amd64" ]; then
    APPIMAGE_ARCH="x86_64"
elif [ "${ARCH}" = "arm64" ]; then
    APPIMAGE_ARCH="aarch64"
fi

# Download appimagetool if not present
APPIMAGETOOL_ARCH="x86_64"
if [ "${ARCH}" = "arm64" ]; then
    APPIMAGETOOL_ARCH="aarch64"
fi
echo "Determining appimagetool architecture..."
echo "ARCH variable: ${ARCH}"
echo "APPIMAGETOOL_ARCH will be: ${APPIMAGETOOL_ARCH}"
APPIMAGETOOL="build/appimagetool-${APPIMAGETOOL_ARCH}.AppImage"

# Clean up wrong architecture appimagetool if exists
if [ "${APPIMAGETOOL_ARCH}" = "aarch64" ] && [ -f "build/appimagetool-x86_64.AppImage" ]; then
    echo "Removing x86_64 appimagetool (need aarch64)..."
    rm -f build/appimagetool-x86_64.AppImage
elif [ "${APPIMAGETOOL_ARCH}" = "x86_64" ] && [ -f "build/appimagetool-aarch64.AppImage" ]; then
    echo "Removing aarch64 appimagetool (need x86_64)..."
    rm -f build/appimagetool-aarch64.AppImage
fi

if [ ! -f "${APPIMAGETOOL}" ]; then
    echo "Downloading appimagetool for ${APPIMAGETOOL_ARCH}..."
    APPIMAGETOOL_URL="https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-${APPIMAGETOOL_ARCH}.AppImage"
    echo "URL: ${APPIMAGETOOL_URL}"
    if ! wget -q "${APPIMAGETOOL_URL}" -O "${APPIMAGETOOL}"; then
        echo "Error: Failed to download appimagetool for ${APPIMAGETOOL_ARCH}"
        echo "Please download it manually from: https://github.com/AppImage/AppImageKit/releases"
        exit 1
    fi
    chmod +x "${APPIMAGETOOL}"
    echo "Downloaded and made executable: ${APPIMAGETOOL}"
    # Verify the architecture of downloaded file
    file "${APPIMAGETOOL}" || true
fi

# Verify appimagetool is executable
if [ ! -x "${APPIMAGETOOL}" ]; then
    echo "Warning: appimagetool is not executable, attempting to fix permissions..."
    if ! chmod +x "${APPIMAGETOOL}"; then
        echo "Error: Failed to make appimagetool executable"
        echo "Please check file permissions on: ${APPIMAGETOOL}"
        exit 1
    fi
fi

# Create AppImage
echo "Creating AppImage..."
echo "Current architecture: ${ARCH}"
echo "AppImage tool architecture: ${APPIMAGETOOL_ARCH}"
echo "System architecture: $(uname -m)"
rm -f "${BUILD_DIR}/${APPIMAGE_NAME}"

# Verify binary architecture before creating AppImage
echo "Verifying binary architecture..."
BINARY_ARCH=$(file "${APPDIR}/usr/bin/${APP_NAME}" | grep -o 'aarch64\|x86-64\|ARM aarch64' || true)
echo "Binary architecture: ${BINARY_ARCH}"
echo "Expected AppImage architecture: ${APPIMAGE_ARCH}"

# Debug: List ALL files in AppDir with their architecture
echo "Listing all ELF files in AppDir:"
find "${APPDIR}" -type f -exec file {} \; 2>/dev/null | grep -E 'ELF|executable|shared' || echo "No ELF files found (besides binary)"

if [ -n "${CI}" ] || ! [ -e /dev/fuse ]; then
    echo "FUSE not available, using --appimage-extract-and-run mode"
    echo "Running: ARCH=${APPIMAGE_ARCH} ${APPIMAGETOOL} --appimage-extract-and-run --verbose ${APPDIR} ${BUILD_DIR}/${APPIMAGE_NAME}"
    if ! ARCH="${APPIMAGE_ARCH}" "${APPIMAGETOOL}" --appimage-extract-and-run --no-appstream --verbose "${APPDIR}" "${BUILD_DIR}/${APPIMAGE_NAME}" 2>&1; then
        echo "Error: AppImage creation failed"
        echo "This might be due to architecture mismatch or missing dependencies"
        echo "Checking AppDir contents..."
        find "${APPDIR}" -type f -exec file {} \; | grep -E 'ELF|shared object' || true

        # Fallback: create tar.gz instead
        echo "Falling back to tar.gz archive..."
        if [ -f "${APPDIR}/usr/bin/${APP_NAME}" ]; then
            tar czf "${BUILD_DIR}/${APP_NAME}-${VERSION}-linux-${ARCH}.tar.gz" -C "${APPDIR}/usr/bin" "${APP_NAME}"
            echo "Created fallback archive: ${BUILD_DIR}/${APP_NAME}-${VERSION}-linux-${ARCH}.tar.gz"
        else
            echo "Error: Binary not found at ${APPDIR}/usr/bin/${APP_NAME}"
            exit 1
        fi
        exit 0
    fi
else
    echo "Running: ARCH=${APPIMAGE_ARCH} ${APPIMAGETOOL} --verbose ${APPDIR} ${BUILD_DIR}/${APPIMAGE_NAME}"
    if ! ARCH="${APPIMAGE_ARCH}" "${APPIMAGETOOL}" --no-appstream --verbose "${APPDIR}" "${BUILD_DIR}/${APPIMAGE_NAME}" 2>&1; then
        echo "Error: AppImage creation failed"
        echo "Checking AppDir contents..."
        find "${APPDIR}" -type f -exec file {} \; | grep -E 'ELF|shared object' || true

        # Fallback: create tar.gz instead
        echo "Falling back to tar.gz archive..."
        if [ -f "${APPDIR}/usr/bin/${APP_NAME}" ]; then
            tar czf "${BUILD_DIR}/${APP_NAME}-${VERSION}-linux-${ARCH}.tar.gz" -C "${APPDIR}/usr/bin" "${APP_NAME}"
            echo "Created fallback archive: ${BUILD_DIR}/${APP_NAME}-${VERSION}-linux-${ARCH}.tar.gz"
        else
            echo "Error: Binary not found at ${APPDIR}/usr/bin/${APP_NAME}"
            exit 1
        fi
        exit 0
    fi
fi

# Clean up
rm -rf "build/appimage"

echo "AppImage created successfully: ${BUILD_DIR}/${APPIMAGE_NAME}"
echo ""
echo "Installation instructions:"
echo "1. Make the AppImage executable: chmod +x ${APPIMAGE_NAME}"
echo "2. Run the AppImage: ./${APPIMAGE_NAME}"
echo ""
echo "User data will be stored in: ~/.local/share/MrRSS/"
