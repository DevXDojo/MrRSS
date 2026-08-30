#!/bin/zsh

set -euo pipefail

SCRIPT_DIR="${0:A:h}"
PROJECT_DIR="${SCRIPT_DIR:h}"
VERSION="${1:-dev}"
DIST_DIR="${SCRIPT_DIR}/dist"
APP_DIR="${DIST_DIR}/MrRSS.app"
CONTENTS_DIR="${APP_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"
SWIFT_BUILD_DIR="${SCRIPT_DIR}/.build/release-universal"
BACKEND_BUILD_DIR="${SCRIPT_DIR}/.build/backend-universal"
CLANG_CACHE_DIR="${SCRIPT_DIR}/.build/clang-module-cache"
DMG_PATH="${DIST_DIR}/MrRSS-${VERSION}-macos.dmg"
MRRSS_ARCH_STRING="${MRRSS_BUILD_ARCHS:-arm64 x86_64}"
MRRSS_ARCHES=("${(@s: :)MRRSS_ARCH_STRING}")
SWIFT_ARCH_ARGUMENTS=()
MRRSS_BUNDLE_VERSION="${VERSION%%-*}"

for MRRSS_ARCH in "${MRRSS_ARCHES[@]}"; do
    SWIFT_ARCH_ARGUMENTS+=(--arch "${MRRSS_ARCH}")
done

mkdir -p "${DIST_DIR}" "${SWIFT_BUILD_DIR}" "${BACKEND_BUILD_DIR}" "${CLANG_CACHE_DIR}"
rm -rf "${APP_DIR}" "${DMG_PATH}"
mkdir -p "${MACOS_DIR}" "${RESOURCES_DIR}"

echo "Building the universal SwiftUI executable..."
cd "${SCRIPT_DIR}"
export CLANG_MODULE_CACHE_PATH="${CLANG_CACHE_DIR}"
swift build \
    --disable-sandbox \
    -c release \
    "${SWIFT_ARCH_ARGUMENTS[@]}" \
    --scratch-path "${SWIFT_BUILD_DIR}"

# A build for several architectures lands under apple/Products/Release, while a
# build for one lands under the triple. Ask for the path rather than assuming.
SWIFT_BIN_DIR="$(swift build \
    --show-bin-path \
    -c release \
    "${SWIFT_ARCH_ARGUMENTS[@]}" \
    --scratch-path "${SWIFT_BUILD_DIR}")"

if [[ ! -x "${SWIFT_BIN_DIR}/MrRSS" ]]; then
    echo "The client executable was not found in ${SWIFT_BIN_DIR}." >&2
    exit 1
fi

cp "${SWIFT_BIN_DIR}/MrRSS" "${MACOS_DIR}/MrRSS"

echo "Building the Go backend for ${MRRSS_ARCH_STRING}..."
cd "${PROJECT_DIR}"
MRRSS_BACKEND_BINARIES=()
for MRRSS_ARCH in "${MRRSS_ARCHES[@]}"; do
    if [[ "${MRRSS_ARCH}" == "x86_64" ]]; then
        MRRSS_GO_ARCH="amd64"
    else
        MRRSS_GO_ARCH="${MRRSS_ARCH}"
    fi
    MRRSS_BACKEND_PATH="${BACKEND_BUILD_DIR}/mrrss-server-${MRRSS_ARCH}"
    CGO_ENABLED=0 GOOS=darwin GOARCH="${MRRSS_GO_ARCH}" go build \
        -trimpath \
        -ldflags="-s -w" \
        -o "${MRRSS_BACKEND_PATH}" .
    MRRSS_BACKEND_BINARIES+=("${MRRSS_BACKEND_PATH}")
done

if (( ${#MRRSS_BACKEND_BINARIES[@]} == 1 )); then
    cp "${MRRSS_BACKEND_BINARIES[1]}" "${RESOURCES_DIR}/mrrss-server"
else
    lipo -create "${MRRSS_BACKEND_BINARIES[@]}" -output "${RESOURCES_DIR}/mrrss-server"
fi
chmod +x "${MACOS_DIR}/MrRSS" "${RESOURCES_DIR}/mrrss-server"

cp "${PROJECT_DIR}/build/darwin/icons.icns" "${RESOURCES_DIR}/AppIcon.icns"

sed \
    -e "s/@VERSION@/${VERSION}/g" \
    -e "s/@BUILD_VERSION@/${MRRSS_BUNDLE_VERSION}/g" \
    "${SCRIPT_DIR}/packaging/Info.plist" > "${CONTENTS_DIR}/Info.plist"

codesign --force --sign - "${RESOURCES_DIR}/mrrss-server"
codesign --force --deep --sign - "${APP_DIR}"

echo "Creating ${DMG_PATH:t}..."
hdiutil create \
    -volname "MrRSS SwiftUI" \
    -srcfolder "${APP_DIR}" \
    -ov \
    -format UDZO \
    "${DMG_PATH}"

echo "Created ${APP_DIR}"
echo "Created ${DMG_PATH}"
