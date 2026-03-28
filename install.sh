#!/bin/sh
# ABOUTME: PVM installer script — detects platform, downloads latest release, installs binary
# ABOUTME: Hosted at https://pvm.tools/install.sh for curl-pipe-sh installation

set -eu

REPO="perigrin/pvm"
INSTALL_DIR="${PVM_INSTALL_DIR:-$HOME/.local/bin}"

main() {
    detect_platform
    fetch_latest_release
    download_and_install
    print_success
}

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)  PLATFORM_OS="linux" ;;
        darwin) PLATFORM_OS="darwin" ;;
        *)
            echo "Error: unsupported operating system: $OS" >&2
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)  PLATFORM_ARCH="amd64" ;;
        aarch64|arm64) PLATFORM_ARCH="arm64" ;;
        *)
            echo "Error: unsupported architecture: $ARCH" >&2
            exit 1
            ;;
    esac

    PLATFORM="${PLATFORM_OS}-${PLATFORM_ARCH}"
    echo "Detected platform: $PLATFORM"
}

fetch_latest_release() {
    echo "Fetching latest release..."
    RELEASE_JSON=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases?per_page=1")

    TAG=$(echo "$RELEASE_JSON" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
    VERSION="${TAG#v}"

    if [ -z "$TAG" ]; then
        echo "Error: could not determine latest release" >&2
        exit 1
    fi

    echo "Latest release: $TAG"
}

download_and_install() {
    ASSET_NAME="pvm-${VERSION}-${PLATFORM}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET_NAME}"

    TMPDIR=$(mktemp -d)
    trap 'rm -rf "$TMPDIR"' EXIT

    echo "Downloading $ASSET_NAME..."
    curl -fsSL -o "${TMPDIR}/${ASSET_NAME}" "$DOWNLOAD_URL"

    echo "Extracting..."
    tar -xzf "${TMPDIR}/${ASSET_NAME}" -C "$TMPDIR"

    mkdir -p "$INSTALL_DIR"
    mv "${TMPDIR}/pvm-${PLATFORM}" "${INSTALL_DIR}/pvm"
    chmod +x "${INSTALL_DIR}/pvm"

    echo "Installed pvm to ${INSTALL_DIR}/pvm"
}

print_success() {
    echo ""
    echo "PVM $VERSION installed successfully."
    echo ""

    # Check if install dir is in PATH
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) ;;
        *)
            echo "Add ${INSTALL_DIR} to your PATH:"
            echo ""
            echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
            echo ""
            ;;
    esac

    echo "Then add shell integration to your profile:"
    echo ""
    echo "  # bash (~/.bashrc) or zsh (~/.zshrc)"
    echo "  eval \"\$(pvm init)\""
    echo ""
    echo "  # fish (~/.config/fish/config.fish)"
    echo "  pvm init | source"
    echo ""
}

main
