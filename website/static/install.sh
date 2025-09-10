#!/usr/bin/env sh

set -eu

GITHUB_ORG="mkideal"
GITHUB_REPO="next"

_cmd=$0
case "$_cmd" in
    *index.sh) ;;
    *) _cmd="sh -s --" ;;
esac

# Check if the terminal supports colors
supports_color() {
    if [ -t 1 ]; then
        ncolors=$(tput colors)
        if [ -n "$ncolors" ] && [ "$ncolors" -ge 8 ]; then
            return 0
        fi
    fi
    return 1
}

# Set color variables based on terminal support
if supports_color; then
    RED=$(tput setaf 1)
    GREEN=$(tput setaf 2)
    YELLOW=$(tput setaf 3)
    BLUE=$(tput setaf 4)
    MAGENTA=$(tput setaf 5)
    CYAN=$(tput setaf 6)
    BOLD=$(tput bold)
    NC=$(tput sgr0) # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    MAGENTA=''
    CYAN=''
    BOLD=''
    NC=''
fi

# Print formatted messages
print_step() {
    echo
    echo "${GREEN}${BOLD}$1${NC}"
}

print_sub_step() {
    echo "  $1"
}

align() {
    local _indent=$1
    local _text=$2
    local cols=$(tput cols)
    if [ "$cols" -gt 120 ]; then
        cols=120
    fi
    echo "$_text" | fold -s -w $((cols - _indent)) | sed -e "2,\$s/^/$(printf '%*s' $_indent '')/"
}

# Print success message
success() {
    local _msg="$*"
    align 0 "${GREEN}$_msg${NC}"
}

# Print information message
info() {
    local _msg="$*"
    align 0 "$_msg"
}

# Print warning message
warn() {
    local _prefix="${YELLOW}Warning: ${NC}"
    local _prefix_length=9  # Length of "Warning: " without color codes
    local _msg="$*"
    printf "$_prefix"
    align $_prefix_length "$_msg"
}

# Print error message
error() {
    local prefix="${RED}${BOLD}Error: ${NC}"
    local prefix_length=7  # Length of "Error: " without color codes
    local msg="$*"
    printf "$prefix"
    align $prefix_length "$msg"
}

die() {
    error $*
    exit 1
}

# Initialize variables
PREFIX=""
VERSION=""
BIN_DIR=""
CACHE_DIR=""
SKIP_EXISTING=0

# Function to print help message
print_help() {
    cat << EOF
Usage: $_cmd [Options]

Options:
  --prefix=PREFIX    Specify the installation prefix (must be an absolute path)
  --version=VERSION  Specify the version to install
  --cache-dir=DIR    Specify the cache directory to store downloaded files
    --skip-existing    If the archive already exists in cache-dir, skip downloading
  -h, --help         Display this help message

Example:
  $_cmd --prefix=/usr/local --version=0.1.0 -i

By default, the script installs the latest version of Next to:
  \$HOME/AppData/Local/Microsoft/WindowsApps (on Windows) 
  \$HOME/.local/bin (on other systems, if \$HOME/.local/bin is in PATH)
  \$HOME/bin (on other systems, if \$HOME/.local/bin is not in PATH)
EOF
}

# Parse command-line arguments
while [ $# -gt 0 ]; do
    case "$1" in
        --prefix=*)
            PREFIX="${1#--prefix=}"
            # Check if PREFIX is an absolute path
            case "$PREFIX" in
                /*) ;;
                *)
                    die "PREFIX must be an absolute path. Received: $PREFIX"
                    ;;
            esac
            shift
            ;;
        --version=*)
            VERSION="${1#--version=}"
            shift
            ;;
        --cache-dir=*)
            CACHE_DIR="${1#--cache-dir=}"
            shift
            ;;
        --skip-existing)
            SKIP_EXISTING=1
            shift
            ;;
        -h|--help)
            print_help
            exit 0
            ;;
        *)
            error "unknown option $1"
            echo
            print_help
            exit 1
            ;;
    esac
done

# Detect OS and architecture
detect_os_arch() {
    print_step "Detecting system information"
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        i386|i486|i586|i686|x86) ARCH="386" ;;
        *) die "Unsupported architecture: $ARCH" ;;
    esac
    case $OS in
        darwin)
            if [ "$ARCH" = "386" ] ; then
                die "32-bit systems are not supported for macOS"
            fi
        ;;
        mingw*) OS="windows" ;;
        linux) ;;
        *) die "Unsupported operating system: $OS" ;;
    esac
    print_sub_step "OS: ${BOLD}$OS${NC}"
    print_sub_step "Architecture: ${BOLD}$ARCH${NC}"
}

# Set default binary and config directories based on OS
set_default_dirs() {
    print_step "Setting up installation directories"
    if [ -z "$PREFIX" ]; then
        if [ "$OS" = "windows" ]; then
            PREFIX="$HOME/AppData/Local/Microsoft/WindowsApps"
            BIN_DIR="$PREFIX"
        else
            case $PATH in
                *":$HOME/.local/bin"|"$HOME/.local/bin:"*|*":$HOME/.local/bin:"*) PREFIX="$HOME/.local" ;;
                *) PREFIX="$HOME" ;;
            esac
        fi
    fi
    if [ -z "$BIN_DIR" ]; then
        BIN_DIR="$PREFIX/bin"
    fi
    print_sub_step "Binary directory: ${BOLD}$BIN_DIR${NC}"
}

# Get the latest stable version
get_latest_version() {
    local _url="https://api.github.com/repos/${GITHUB_ORG}/${GITHUB_REPO}/releases/latest"
    print_step "Fetching latest version information"
    print_sub_step "URL: $_url"
    LATEST_VERSION=$(curl -sSf $_url | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST_VERSION" ]; then
        die "Failed to get the latest version. Please check your internet connection or try again later."
    fi
    LATEST_VERSION=${LATEST_VERSION#v} # Remove 'v' prefix if present
    print_sub_step "Latest: ${BOLD}$LATEST_VERSION${NC}"
}

clear_last_line() {
    printf "\033[1A\033[2K"
}

# Download the appropriate Next package
download_next() {
    VERSION=${VERSION:-$LATEST_VERSION}
    FILENAME="next$VERSION.$OS-$ARCH.tar.gz"
    if [ "$OS" = "windows" ]; then
        FILENAME="next$VERSION.$OS-$ARCH.zip"
    fi
    URL="https://github.com/${GITHUB_ORG}/${GITHUB_REPO}/releases/download/v$VERSION/$FILENAME"
    print_step "Downloading $FILENAME"
    print_sub_step "URL: $URL"

    if [ -n "$CACHE_DIR" ]; then
        mkdir -p "$CACHE_DIR" || die "Failed to create cache dir"
        ARCHIVE_PATH="$CACHE_DIR/$FILENAME"
        echo "Archive path: $ARCHIVE_PATH, SKIP_EXISTING: $SKIP_EXISTING"
        if [ -f "$ARCHIVE_PATH" ] && [ $SKIP_EXISTING -eq 1 ]; then
            print_sub_step "Using cached file (skip-existing)"
        else
            TMP_DL="$ARCHIVE_PATH.tmp"
            if ! curl -fSL --progress-bar "$URL" -o "$TMP_DL"; then
                rm -f "$TMP_DL"
                die "Failed to download Next. Please check your internet connection and try again."
            fi
            mv "$TMP_DL" "$ARCHIVE_PATH"
            clear_last_line
        fi
        TEMP_DIR=$(mktemp -d) || die "Failed to create temporary directory"
        ARCHIVE="$ARCHIVE_PATH"
        DOWNLOAD_DIR="$TEMP_DIR"
    else
        TEMP_DIR=$(mktemp -d) || die "Failed to create temporary directory"
        if ! curl -fSL --progress-bar "$URL" -o "$TEMP_DIR/$FILENAME"; then
            rm -rf "$TEMP_DIR"
            die "Failed to download Next. Please check your internet connection and try again."
        fi
        clear_last_line
        ARCHIVE="$TEMP_DIR/$FILENAME"
        DOWNLOAD_DIR="$TEMP_DIR"
    fi
}

# Install Next
install_next() {
    print_step "Installing to $BIN_DIR"
    print_sub_step "Extracting $FILENAME"
    if [ "$OS" = "windows" ]; then
    if ! unzip -q "$ARCHIVE" -d "$DOWNLOAD_DIR"; then
            rm -rf "$DOWNLOAD_DIR"
            die "Failed to extract Next package."
        fi
    else
    if ! tar -xzf "$ARCHIVE" -C "$DOWNLOAD_DIR"; then
            rm -rf "$DOWNLOAD_DIR"
            die "Failed to extract Next package."
        fi
    fi

    mkdir -p "$BIN_DIR" || die "Failed to create installation directory $BIN_DIR"

    print_sub_step "Copying binaries to $BIN_DIR"
    mv "$DOWNLOAD_DIR/next$VERSION.$OS-$ARCH/bin/"* "$BIN_DIR/" || die "Failed to install Next binary."

    # Clean up the temporary directory
    rm -rf "$DOWNLOAD_DIR"

    print_step "Next has been successfully installed!"
}

# Check if the installation directory is in PATH
check_path() {
    if ! echo "$PATH" | tr ':' '\n' | grep -qx "$BIN_DIR"; then
        warn "Installation directory is not in your PATH."
        info "Add the following line to your shell configuration file (.bashrc, .zshrc, etc.):"
        info "${MAGENTA}export PATH=\"\$PATH:$BIN_DIR\"${NC}"
    fi
}

# Display countdown
countdown() {
    local _fmt="${YELLOW}Installation will start in ${BOLD}%d${NC}${YELLOW} seconds. Press Ctrl+C to cancel.${NC}"
    printf "${_fmt}" 5
    for i in 4 3 2 1; do
        sleep 1
        printf "\r${_fmt}" $i
    done
    printf "\r%*s\r" $(tput cols) ""
}

# Print welcome message in a box
print_welcome() {
    message="Welcome to the Next Installer"
    padding="  "
    width=$(( $(printf "%s" "$message" | wc -c) + $(printf "%s" "$padding" | wc -c) * 2 ))

    print_line() {
        i=1
        while [ $i -le "$1" ]; do
            printf "%s" "─"
            i=$((i + 1))
        done
    }

    printf '╭'
    print_line "$width"
    printf '╮\n'
    
    printf '│%s%b%s%b%s│\n' "$padding" "$BOLD" "$message" "$NC" "$padding"
    
    printf '╰'
    print_line "$width"
    printf '╯'
    printf '%b\n' "$NC"
}

# Main installation process
main() {
    print_welcome

    detect_os_arch
    set_default_dirs

    if [ -z "$VERSION" ]; then
        get_latest_version
    else
        LATEST_VERSION="$VERSION"
    fi

    download_next
    install_next
    check_path

    info
    info "Run ${BOLD}${MAGENTA}next -h${NC} to get started or run ${BOLD}${MAGENTA}next version${NC} to check the installed version."
}

# Run the installation
main
