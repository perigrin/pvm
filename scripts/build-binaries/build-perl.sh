#!/bin/bash
# ABOUTME: Main script for building Perl binaries across different platforms
# ABOUTME: Handles cross-platform compilation, packaging, and checksum generation

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Default values
VERSION=""
PLATFORM=""
OUTPUT_DIR="${PROJECT_ROOT}/build/binaries"
BUILD_CONFIG="${SCRIPT_DIR}/build-config.yaml"
PLATFORMS_CONFIG="${SCRIPT_DIR}/platforms.json"
VERBOSE=false
CLEAN=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build Perl binaries for distribution.

OPTIONS:
    -v, --version VERSION     Perl version to build (required)
    -p, --platform PLATFORM  Target platform (e.g., linux-amd64, darwin-arm64)
    -o, --output DIR         Output directory (default: ${OUTPUT_DIR})
    -c, --config FILE        Build configuration file (default: ${BUILD_CONFIG})
    --platforms FILE         Platforms configuration file (default: ${PLATFORMS_CONFIG})
    --clean                  Clean build directory before building
    --verbose                Enable verbose output
    -h, --help               Show this help message

EXAMPLES:
    $0 --version 5.38.2 --platform linux-amd64
    $0 --version 5.38.2 --platform darwin-arm64 --clean
    $0 --version 5.38.2 --platform all  # Build for all supported platforms

SUPPORTED PLATFORMS:
    linux-amd64, linux-arm64
    darwin-amd64, darwin-arm64
    windows-amd64
EOF
}

log() {
    local level="$1"
    shift
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    case "$level" in
        INFO)  echo -e "${GREEN}[INFO]${NC}  ${timestamp} $*" ;;
        WARN)  echo -e "${YELLOW}[WARN]${NC}  ${timestamp} $*" ;;
        ERROR) echo -e "${RED}[ERROR]${NC} ${timestamp} $*" >&2 ;;
        DEBUG) [[ "$VERBOSE" == "true" ]] && echo -e "${BLUE}[DEBUG]${NC} ${timestamp} $*" ;;
    esac
}

check_dependencies() {
    log INFO "Checking build dependencies..."

    local missing_deps=()

    # Check for required tools
    command -v curl >/dev/null 2>&1 || missing_deps+=(curl)
    command -v tar >/dev/null 2>&1 || missing_deps+=(tar)
    command -v make >/dev/null 2>&1 || missing_deps+=(make)
    command -v gcc >/dev/null 2>&1 || missing_deps+=(gcc)
    command -v jq >/dev/null 2>&1 || missing_deps+=(jq)

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log ERROR "Missing required dependencies: ${missing_deps[*]}"
        log ERROR "Please install the missing dependencies and try again."
        exit 1
    fi

    log INFO "All dependencies satisfied."
}

load_platform_config() {
    local platform="$1"

    if [[ ! -f "$PLATFORMS_CONFIG" ]]; then
        log ERROR "Platforms configuration file not found: $PLATFORMS_CONFIG"
        exit 1
    fi

    # Parse platform configuration using jq
    local config
    config=$(jq -r ".platforms.\"$platform\"" "$PLATFORMS_CONFIG")

    if [[ "$config" == "null" ]]; then
        log ERROR "Unsupported platform: $platform"
        log ERROR "Check $PLATFORMS_CONFIG for supported platforms."
        exit 1
    fi

    # Export platform-specific variables
    export PLATFORM_OS=$(echo "$config" | jq -r '.os')
    export PLATFORM_ARCH=$(echo "$config" | jq -r '.arch')
    export PLATFORM_CC=$(echo "$config" | jq -r '.cc // "gcc"')
    export PLATFORM_CFLAGS=$(echo "$config" | jq -r '.cflags // ""')
    export PLATFORM_ARCHIVE_EXT=$(echo "$config" | jq -r '.archive_ext')
    export PLATFORM_BINARY_EXT=$(echo "$config" | jq -r '.binary_ext // ""')

    log DEBUG "Platform config loaded: os=$PLATFORM_OS, arch=$PLATFORM_ARCH"
}

download_perl_source() {
    local version="$1"
    local build_dir="$2"

    log INFO "Downloading Perl $version source..."

    local perl_url="https://www.cpan.org/src/5.0/perl-${version}.tar.gz"
    local source_archive="${build_dir}/perl-${version}.tar.gz"
    local source_dir="${build_dir}/perl-${version}"

    # Download source if not already present
    if [[ ! -f "$source_archive" ]]; then
        log INFO "Downloading from: $perl_url"
        curl -fsSL "$perl_url" -o "$source_archive" || {
            log ERROR "Failed to download Perl source"
            exit 1
        }
    else
        log INFO "Using cached source archive: $source_archive"
    fi

    # Extract source
    if [[ ! -d "$source_dir" ]]; then
        log INFO "Extracting Perl source..."
        tar -xzf "$source_archive" -C "$build_dir"
    else
        log INFO "Using existing source directory: $source_dir"
    fi

    echo "$source_dir"
}

configure_perl_build() {
    local source_dir="$1"
    local install_prefix="$2"

    log INFO "Configuring Perl build..."

    cd "$source_dir"

    # Load build configuration
    local configure_opts=""
    if [[ -f "$BUILD_CONFIG" ]]; then
        # Parse YAML configuration (simplified - real implementation would use a proper YAML parser)
        configure_opts=$(grep -E "^\s*configure_opts:" "$BUILD_CONFIG" | sed 's/.*configure_opts:\s*//' || echo "")
    fi

    # Default configure options for binary distribution
    local default_opts=(
        "-des"  # Use defaults for all questions
        "-Dprefix=$install_prefix"
        "-Dusethreads"
        "-Dusemultiplicity"
        "-Duseperlio"
        "-Duse64bitall"
        "-Doptimize=-O2"
    )

    # Platform-specific configuration
    case "$PLATFORM_OS" in
        linux)
            default_opts+=("-Dcc=$PLATFORM_CC")
            [[ -n "$PLATFORM_CFLAGS" ]] && default_opts+=("-Dccflags=$PLATFORM_CFLAGS")
            ;;
        darwin)
            default_opts+=("-Dcc=$PLATFORM_CC")
            default_opts+=("-Dld=$PLATFORM_CC")
            ;;
        windows)
            # Windows-specific configuration would go here
            log WARN "Windows build configuration not fully implemented"
            ;;
    esac

    log INFO "Running Configure with options: ${default_opts[*]} $configure_opts"

    # Run Configure
    ./Configure "${default_opts[@]}" $configure_opts || {
        log ERROR "Perl Configure failed"
        exit 1
    }
}

build_perl() {
    local source_dir="$1"

    log INFO "Building Perl..."

    cd "$source_dir"

    # Build Perl
    local jobs=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "1")
    log INFO "Building with $jobs parallel jobs"

    make -j"$jobs" || {
        log ERROR "Perl build failed"
        exit 1
    }

    # Run tests (optional, can be disabled for faster builds)
    if [[ "${RUN_TESTS:-false}" == "true" ]]; then
        log INFO "Running Perl tests..."
        make test || {
            log WARN "Some tests failed, but continuing with packaging"
        }
    fi

    log INFO "Installing Perl..."
    make install || {
        log ERROR "Perl installation failed"
        exit 1
    }
}

create_binary_package() {
    local install_dir="$1"
    local version="$2"
    local platform="$3"
    local output_dir="$4"

    log INFO "Creating binary package..."

    local package_name="perl-${version}-${platform}"
    local package_dir="${output_dir}/${package_name}"
    local archive_name="${package_name}.${PLATFORM_ARCHIVE_EXT}"
    local archive_path="${output_dir}/${archive_name}"

    # Create package directory structure
    mkdir -p "$package_dir"

    # Copy installation files
    log INFO "Copying installation files..."
    cp -r "$install_dir"/* "$package_dir/"

    # Create metadata file
    cat > "${package_dir}/PERL_BUILD_INFO" << EOF
# Perl Binary Distribution Information
Version: $version
Platform: $platform
OS: $PLATFORM_OS
Architecture: $PLATFORM_ARCH
Build Date: $(date -u '+%Y-%m-%d %H:%M:%S UTC')
Built By: $(whoami)@$(hostname)
Build Script: $0
Compiler: $PLATFORM_CC
EOF

    # Create archive
    log INFO "Creating archive: $archive_name"
    cd "$output_dir"

    case "$PLATFORM_ARCHIVE_EXT" in
        tar.gz)
            tar -czf "$archive_name" "${package_name}/"
            ;;
        zip)
            zip -r "$archive_name" "${package_name}/"
            ;;
        *)
            log ERROR "Unsupported archive format: $PLATFORM_ARCHIVE_EXT"
            exit 1
            ;;
    esac

    # Generate checksum
    log INFO "Generating checksum..."
    local checksum_file="${archive_path}.sha256"

    case "$(uname)" in
        Linux)   sha256sum "$archive_name" > "$checksum_file" ;;
        Darwin)  shasum -a 256 "$archive_name" > "$checksum_file" ;;
        *)       log WARN "Cannot generate checksum on this platform" ;;
    esac

    # Clean up package directory
    rm -rf "$package_dir"

    log INFO "Binary package created: $archive_path"
    log INFO "Checksum file: $checksum_file"

    # Display package information
    local size=$(ls -lh "$archive_path" | awk '{print $5}')
    log INFO "Package size: $size"

    if [[ -f "$checksum_file" ]]; then
        local checksum=$(cut -d' ' -f1 "$checksum_file")
        log INFO "SHA256: $checksum"
    fi
}

build_platform() {
    local version="$1"
    local platform="$2"

    log INFO "Building Perl $version for $platform"

    # Load platform configuration
    load_platform_config "$platform"

    # Create build directories
    local build_base="${OUTPUT_DIR}/build"
    local build_dir="${build_base}/${platform}"
    local install_dir="${build_dir}/install"

    mkdir -p "$build_dir" "$install_dir"

    # Download and extract Perl source
    local source_dir
    source_dir=$(download_perl_source "$version" "$build_dir")

    # Configure Perl build
    configure_perl_build "$source_dir" "$install_dir"

    # Build and install Perl
    build_perl "$source_dir"

    # Create binary package
    create_binary_package "$install_dir" "$version" "$platform" "$OUTPUT_DIR"

    log INFO "Successfully built Perl $version for $platform"
}

main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -p|--platform)
                PLATFORM="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -c|--config)
                BUILD_CONFIG="$2"
                shift 2
                ;;
            --platforms)
                PLATFORMS_CONFIG="$2"
                shift 2
                ;;
            --clean)
                CLEAN=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log ERROR "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    # Validate required arguments
    if [[ -z "$VERSION" ]]; then
        log ERROR "Perl version is required (use --version)"
        usage
        exit 1
    fi

    if [[ -z "$PLATFORM" ]]; then
        log ERROR "Platform is required (use --platform)"
        usage
        exit 1
    fi

    # Check dependencies
    check_dependencies

    # Clean output directory if requested
    if [[ "$CLEAN" == "true" ]]; then
        log INFO "Cleaning output directory: $OUTPUT_DIR"
        rm -rf "$OUTPUT_DIR"
    fi

    # Create output directory
    mkdir -p "$OUTPUT_DIR"

    # Build for specified platform(s)
    if [[ "$PLATFORM" == "all" ]]; then
        # Build for all supported platforms
        if [[ ! -f "$PLATFORMS_CONFIG" ]]; then
            log ERROR "Platforms configuration file not found: $PLATFORMS_CONFIG"
            exit 1
        fi

        local platforms
        platforms=$(jq -r '.platforms | keys[]' "$PLATFORMS_CONFIG")

        for platform in $platforms; do
            log INFO "Starting build for platform: $platform"
            build_platform "$VERSION" "$platform"
        done
    else
        # Build for single platform
        build_platform "$VERSION" "$PLATFORM"
    fi

    log INFO "Build completed successfully!"
}

# Run main function
main "$@"
