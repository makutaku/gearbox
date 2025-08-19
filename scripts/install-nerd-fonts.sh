#!/bin/bash

# Nerd Fonts Installation Script for Debian Linux
# Patched fonts with icons and glyphs for developers
# Usage: ./install-nerd-fonts.sh [OPTIONS]

# Find the script directory and load common library first
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_DIR="$(dirname "$SCRIPT_DIR")"

# Source common library for shared functions
if [[ -f "$REPO_DIR/lib/common.sh" ]]; then
    source "$REPO_DIR/lib/common.sh"
else
    echo "ERROR: common.sh not found in $REPO_DIR/lib/" >&2
    exit 1
fi

# Enable error handling (without -u to avoid breaking existing libraries)
set -eo pipefail

# Configuration
readonly NERD_FONTS_REPO="https://github.com/ryanoasis/nerd-fonts"
readonly NERD_FONTS_VERSION="v3.1.1"
readonly FONTS_DIR="$HOME/.local/share/fonts"
readonly DOWNLOAD_DIR="$HOME/tools/build/nerd-fonts"

# Available fonts - expanded collection from nerd-fonts repository
readonly -a MINIMAL_FONTS=(
    "FiraCode"
    "JetBrainsMono" 
    "Hack"
)

readonly -a STANDARD_FONTS=(
    "FiraCode"
    "JetBrainsMono"
    "Hack"
    "SourceCodePro"
    "Inconsolata"
    "CascadiaCode"
    "UbuntuMono"
    "DejaVuSansMono"
    "RobotoMono"
    "SpaceMono"
    "Iosevka"
    "GeistMono"
)

readonly -a MAXIMUM_FONTS=(
    # All fonts alphabetically ordered (exact names from nerd-fonts v3.1.1 release)
    "0xProto"
    "3270"
    "Agave"
    "AnonymousPro"
    "Arimo"
    "AurulentSansMono"
    "BigBlueTerminal"
    "BitstreamVeraSansMono"
    "CascadiaCode"
    "CascadiaMono"
    "CodeNewRoman"
    "ComicShannsMono"
    "CommitMono"
    "Cousine"
    "D2Coding"
    "DaddyTimeMono"
    "DejaVuSansMono"
    "DroidSansMono"
    "EnvyCodeR"
    "FantasqueSansMono"
    "FiraCode"
    "FiraMono"
    "GeistMono"
    "Go-Mono"
    "Gohu"
    "Hack"
    "Hasklig"
    "HeavyData"
    "Hermit"
    "iA-Writer"
    "IBMPlexMono"
    "Inconsolata"
    "InconsolataGo"
    "InconsolataLGC"
    "IntelOneMono"
    "Iosevka"
    "IosevkaTerm"
    "IosevkaTermSlab"
    "JetBrainsMono"
    "Lekton"
    "LiberationMono"
    "Lilex"
    "MartianMono"
    "Meslo"
    "Monaspace"
    "Monofur"
    "Monoid"
    "Mononoki"
    "MPlus"
    "NerdFontsSymbolsOnly"
    "Noto"
    "OpenDyslexic"
    "Overpass"
    "ProFont"
    "ProggyClean"
    "RobotoMono"
    "ShareTechMono"
    "SourceCodePro"
    "SpaceMono"
    "Terminus"
    "Tinos"
    "Ubuntu"
    "UbuntuMono"
    "VictorMono"
)

# Default options
BUILD_TYPE="standard"
MODE="install"
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
INTERACTIVE=false
SPECIFIC_FONTS=""
CONFIGURE_APPS=false
PREVIEW=false

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --minimal)
                BUILD_TYPE="minimal"
                shift
                ;;
            --standard)
                BUILD_TYPE="standard"
                shift
                ;;
            --maximum)
                BUILD_TYPE="maximum"
                shift
                ;;
            -c|--config-only)
                MODE="config"
                shift
                ;;
            -b|--build-only)
                MODE="build"
                shift
                ;;
            -i|--install)
                MODE="install"
                shift
                ;;
            --skip-deps)
                SKIP_DEPS=true
                shift
                ;;
            --run-tests)
                RUN_TESTS=true
                shift
                ;;
            --force)
                FORCE_INSTALL=true
                shift
                ;;
            --interactive)
                INTERACTIVE=true
                shift
                ;;
            --fonts=*)
                SPECIFIC_FONTS="${1#*=}"
                shift
                ;;
            --configure-apps)
                CONFIGURE_APPS=true
                shift
                ;;
            --preview)
                PREVIEW=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Show help
show_help() {
    cat << EOF
Patched fonts with icons and glyphs for developers

Usage: $0 [OPTIONS]

Build Types:
  --minimal            Essential fonts (3 fonts: FiraCode, JetBrains Mono, Hack)
  --standard           Popular fonts (12 fonts, ~120MB) - default
  --maximum            All fonts (64 fonts, ~640MB)

Font Selection:
  --fonts="Font1,Font2"    Install specific fonts (comma-separated)
  --interactive        Interactive font selection menu
  --preview            Show font previews before selection

Configuration:
  --configure-apps     Automatically configure VS Code, terminals
  --skip-deps          Skip dependency installation
  --run-tests          Verify fonts after installation
  --force              Force reinstallation if already installed

Examples:
  $0                               # Install standard font collection
  $0 --minimal                     # Fast install (3 essential fonts)
  $0 --fonts="FiraCode,Hack"       # Install specific fonts
  $0 --interactive                 # Choose fonts interactively
EOF
}

# Install dependencies
install_dependencies() {
    if [[ "$SKIP_DEPS" == true ]]; then
        log "Skipping dependency installation"
        return 0
    fi

    log "Installing Nerd Fonts dependencies..."
    
    # Check if running in a container or CI environment
    if [[ -n "${CI:-}" || -n "${CONTAINER:-}" ]]; then
        warning "Running in CI/container environment, skipping interactive package installation"
        return 0
    fi

    # Install required packages
    local -a packages=("fontconfig" "curl" "unzip")
    local -a missing_packages=()
    
    for package in "${packages[@]}"; do
        if ! dpkg -l "$package" &>/dev/null; then
            missing_packages+=("$package")
        fi
    done
    
    if [[ ${#missing_packages[@]} -gt 0 ]]; then
        log "Installing missing packages: ${missing_packages[*]}"
        sudo apt update || { error "Failed to update package list"; return 1; }
        sudo apt install -y "${missing_packages[@]}" || { error "Failed to install packages"; return 1; }
        success "Dependencies installed successfully"
    else
        success "All dependencies are already installed"
    fi
}

# Get font list based on build type
get_font_list() {
    local build_type="$1"
    case "$build_type" in
        "minimal")
            printf '%s\n' "${MINIMAL_FONTS[@]}"
            ;;
        "standard")
            printf '%s\n' "${STANDARD_FONTS[@]}"
            ;;
        "maximum")
            printf '%s\n' "${MAXIMUM_FONTS[@]}"
            ;;
        *)
            error "Unknown build type: $build_type"
            return 1
            ;;
    esac
}

# Parse specific fonts from comma-separated string
parse_specific_fonts() {
    local input="$1"
    local -a font_list=()
    
    IFS=',' read -ra font_array <<< "$input"
    for font in "${font_array[@]}"; do
        font=$(echo "$font" | xargs)  # Trim whitespace
        
        # Check if font exists in maximum collection
        local font_found=false
        for max_font in "${MAXIMUM_FONTS[@]}"; do
            if [[ "$font" == "$max_font" ]]; then
                font_list+=("$font")
                font_found=true
                break
            fi
        done
        
        if [[ "$font_found" == false ]]; then
            warning "Unknown font: $font. Available fonts: ${MAXIMUM_FONTS[*]}"
        fi
    done
    
    printf '%s\n' "${font_list[@]}"
}

# Interactive font selection - simple and robust approach
interactive_font_selection() {
    # Check if we're in an interactive terminal
    if [[ ! -t 0 ]]; then
        echo "🎨 Non-interactive mode detected - using standard font collection" >&2
        printf '%s\n' "${STANDARD_FONTS[@]}"
        return 0
    fi
    
    echo "🎨 Interactive Font Selection" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo >&2
    
    # Simple menu using bash select builtin - much more reliable than custom terminal handling
    local PS3="Choose an option (1-5): "
    select choice in \
        "Quick: Standard fonts (12 fonts, ~120MB) - Recommended" \
        "Quick: Minimal fonts (3 essential fonts, ~30MB)" \
        "Quick: All fonts (64 fonts, ~640MB)" \
        "Custom: Choose specific fonts" \
        "Cancel"; do
        
        case $REPLY in
            1)
                echo "Selected: Standard font collection" >&2
                printf '%s\n' "${STANDARD_FONTS[@]}"
                return 0
                ;;
            2)
                echo "Selected: Minimal font collection" >&2
                printf '%s\n' "${MINIMAL_FONTS[@]}"
                return 0
                ;;
            3)
                echo "Selected: Maximum font collection" >&2
                printf '%s\n' "${MAXIMUM_FONTS[@]}"
                return 0
                ;;
            4)
                break  # Go to custom selection
                ;;
            5)
                echo "Installation cancelled" >&2
                exit 0
                ;;
            *)
                echo "Invalid choice. Please enter 1, 2, 3, 4, or 5." >&2
                ;;
        esac
    done
    
    # Custom font selection - simple numbered approach
    echo >&2
    echo "Custom Font Selection" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo >&2
    echo "Available fonts:" >&2
    
    local -a available_fonts=("${MAXIMUM_FONTS[@]}")
    local -i i=1
    for font in "${available_fonts[@]}"; do
        printf "%2d) %s\n" $i "$font" >&2
        ((i++))
    done
    
    echo >&2
    echo "Enter font numbers separated by spaces or commas (e.g., '1 3 5' or '1,3,5'):" >&2
    echo "Or press ENTER for standard collection:" >&2
    read -r user_input
    
    # Handle empty input (default to standard)
    if [[ -z "$user_input" ]]; then
        echo "Using standard font collection" >&2
        printf '%s\n' "${STANDARD_FONTS[@]}"
        return 0
    fi
    
    # Parse user input
    local -a selected_fonts=()
    # Replace commas with spaces and split
    user_input="${user_input//,/ }"
    for num in $user_input; do
        # Validate it's a number
        if [[ "$num" =~ ^[0-9]+$ ]] && [[ $num -ge 1 ]] && [[ $num -le ${#available_fonts[@]} ]]; then
            local index=$((num - 1))
            selected_fonts+=("${available_fonts[index]}")
        else
            echo "Invalid selection: $num (must be 1-${#available_fonts[@]})" >&2
        fi
    done
    
    if [[ ${#selected_fonts[@]} -eq 0 ]]; then
        echo "No valid fonts selected. Using standard collection." >&2
        printf '%s\n' "${STANDARD_FONTS[@]}"
        return 0
    fi
    
    echo "Selected ${#selected_fonts[@]} fonts: ${selected_fonts[*]}" >&2
    printf '%s\n' "${selected_fonts[@]}"
}

# Check if a font is already installed
is_font_installed() {
    local font="$1"
    
    # Use fc-list to check if font is installed, handle errors gracefully
    if command -v fc-list >/dev/null 2>&1; then
        # Disable error exit temporarily for the grep command
        set +e
        fc-list 2>/dev/null | grep -qi "$font.*nerd" 2>/dev/null
        local grep_result=$?
        set -e
        
        if [[ $grep_result -eq 0 ]]; then
            return 0  # Font is installed
        fi
    fi
    
    return 1  # Font is not installed
}

# Download a font
download_font() {
    local font="$1"
    local font_url="$NERD_FONTS_REPO/releases/download/$NERD_FONTS_VERSION/$font.zip"
    local font_zip="$DOWNLOAD_DIR/$font.zip"
    
    log "  Downloading $font..."
    
    if ! curl -fsSL "$font_url" -o "$font_zip"; then
        error "Failed to download $font from $font_url"
        return 1
    fi
    
    return 0
}

# Extract and install a font
extract_and_install_font() {
    local font="$1"
    local font_zip="$DOWNLOAD_DIR/$font.zip"
    local font_extract_dir="$DOWNLOAD_DIR/$font"
    
    log "  Extracting $font..."
    
    # Clean up any existing extraction directory
    rm -rf "$font_extract_dir"
    mkdir -p "$font_extract_dir"
    
    if ! unzip -q "$font_zip" -d "$font_extract_dir"; then
        error "Failed to extract $font"
        return 1
    fi
    
    log "  Installing font files..."
    
    # Install font files
    local font_count=0
    while IFS= read -r -d '' font_file; do
        if [[ -f "$font_file" ]]; then
            cp "$font_file" "$FONTS_DIR/"
            ((font_count++))
        fi
    done < <(find "$font_extract_dir" \( -name "*.ttf" -o -name "*.otf" \) -print0)
    
    if [[ $font_count -eq 0 ]]; then
        warning "No font files found in $font"
        return 1
    fi
    
    log "  Installed $font_count font files"
    
    # Clean up
    rm -f "$font_zip"
    rm -rf "$font_extract_dir"
    
    return 0
}

# Install a single font
install_font() {
    local font="$1"
    
    log "📦 Installing $font..."
    
    # Check if font is already installed (unless force)
    if [[ "$FORCE_INSTALL" != true ]] && is_font_installed "$font"; then
        success "  $font is already installed (use --force to reinstall)"
        return 0
    fi
    
    # Download font
    if ! download_font "$font"; then
        return 1
    fi
    
    # Extract and install
    if ! extract_and_install_font "$font"; then
        return 1
    fi
    
    success "  $font installed successfully"
    return 0
}

# Install fonts
install_fonts() {
    local -a fonts_to_install=()
    
    # Determine which fonts to install
    if [[ -n "$SPECIFIC_FONTS" ]]; then
        readarray -t fonts_to_install < <(parse_specific_fonts "$SPECIFIC_FONTS")
        log "Installing specific fonts: ${fonts_to_install[*]}"
    elif [[ "$INTERACTIVE" == true ]]; then
        readarray -t fonts_to_install < <(interactive_font_selection)
        log "Installing selected fonts: ${fonts_to_install[*]}"
    else
        readarray -t fonts_to_install < <(get_font_list "$BUILD_TYPE")
        log "Installing $BUILD_TYPE collection: ${fonts_to_install[*]}"
    fi
    
    if [[ ${#fonts_to_install[@]} -eq 0 ]]; then
        error "No fonts selected for installation"
        return 1
    fi
    
    # Create directories
    mkdir -p "$FONTS_DIR" "$DOWNLOAD_DIR"
    
    local total_fonts=${#fonts_to_install[@]}
    local current_font=0
    local failed_fonts=()
    
    log "🎨 Installing Nerd Fonts ($total_fonts fonts)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Install each font
    for font in "${fonts_to_install[@]}"; do
        ((current_font++))
        log "[$current_font/$total_fonts] Processing $font"
        
        if ! install_font "$font"; then
            failed_fonts+=("$font")
            warning "Failed to install $font"
        fi
    done
    
    # Report results
    local successful_fonts=$((total_fonts - ${#failed_fonts[@]}))
    log "Installation complete: $successful_fonts/$total_fonts fonts installed"
    
    if [[ ${#failed_fonts[@]} -gt 0 ]]; then
        warning "Failed fonts: ${failed_fonts[*]}"
    fi
    
    # Refresh font cache
    log "🔄 Refreshing font cache..."
    if fc-cache -fv >/dev/null 2>&1; then
        success "Font cache updated successfully"
    else
        warning "Font cache update failed, but fonts may still work"
    fi
    
    return 0
}

# Verify installation
verify_installation() {
    local -a fonts_to_verify=()
    
    if [[ -n "$SPECIFIC_FONTS" ]]; then
        readarray -t fonts_to_verify < <(parse_specific_fonts "$SPECIFIC_FONTS")
    else
        readarray -t fonts_to_verify < <(get_font_list "$BUILD_TYPE")
    fi
    
    local verified_count=0
    
    log "🔍 Verifying font installation..."
    
    for font in "${fonts_to_verify[@]}"; do
        if is_font_installed "$font"; then
            success "  ✓ $font verified"
            ((verified_count++))
        else
            error "  ✗ $font not found"
        fi
    done
    
    local total_fonts=${#fonts_to_verify[@]}
    if [[ $verified_count -eq $total_fonts ]]; then
        success "All $total_fonts fonts verified successfully!"
    else
        warning "$verified_count/$total_fonts fonts verified"
    fi
    
    # Show installation summary
    echo
    log "📊 Installation Summary"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "✅ %-20s %d fonts\\n" "Installed:" $verified_count
    printf "💾 %-20s %s\\n" "Location:" "$FONTS_DIR"
    printf "📏 %-20s %s\\n" "Disk usage:" "$(du -sh "$FONTS_DIR" 2>/dev/null | cut -f1 || echo "Unknown")"
    echo
    
    return 0
}

# Configure applications
configure_applications() {
    log "⚙️  Configuring applications..."
    
    # VS Code configuration
    if command -v code &>/dev/null; then
        local vscode_settings="$HOME/.config/Code/User/settings.json"
        local vscode_dir="$(dirname "$vscode_settings")"
        
        mkdir -p "$vscode_dir"
        
        if [[ -f "$vscode_settings" ]]; then
            if ! grep -q "editor.fontFamily" "$vscode_settings"; then
                # Add font family setting using jq if available
                if command -v jq &>/dev/null; then
                    local temp_file
                    temp_file=$(mktemp)
                    jq '. + {"editor.fontFamily": "FiraCode Nerd Font", "editor.fontLigatures": true}' "$vscode_settings" > "$temp_file"
                    mv "$temp_file" "$vscode_settings"
                    success "  VS Code configuration updated"
                else
                    warning "  jq not available, VS Code configuration skipped"
                fi
            else
                warning "  VS Code font already configured"
            fi
        else
            # Create new settings file
            cat > "$vscode_settings" << 'EOF'
{
    "editor.fontFamily": "FiraCode Nerd Font",
    "editor.fontLigatures": true,
    "terminal.integrated.fontFamily": "JetBrains Mono Nerd Font"
}
EOF
            success "  VS Code configuration created"
        fi
    fi
    
    return 0
}

# Main execution
main() {
    log "🎨 Nerd Fonts Installation"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Detect system configuration
    log "📋 System: $(lsb_release -d 2>/dev/null | cut -f2 || echo "Unknown Linux")"
    log "📁 Font directory: $FONTS_DIR"
    log "💾 Available space: $(df -h "$HOME" | awk 'NR==2 {print $4}' || echo "Unknown")"
    echo
    
    # Install dependencies
    if [[ "$MODE" != "config" ]]; then
        if ! install_dependencies; then
            error "Failed to install dependencies"
            return 1
        fi
    fi
    
    # Exit early if config-only mode
    if [[ "$MODE" == "config" ]]; then
        success "Configuration completed. Run without --config-only to install fonts."
        return 0
    fi
    
    # Install fonts
    if [[ "$MODE" != "build" ]]; then
        if ! install_fonts; then
            error "Font installation failed"
            return 1
        fi
        
        # Verify installation
        if [[ "$RUN_TESTS" == true ]]; then
            verify_installation
        fi
        
        # Configure applications
        if [[ "$CONFIGURE_APPS" == true ]]; then
            configure_applications
        fi
        
        success "🎉 Nerd Fonts installation completed!"
    else
        success "Build preparation completed. Run without --build-only to install fonts."
    fi
    
    return 0
}

# Entry point
ensure_not_root
parse_arguments "$@"

# Execute main function
if ! main; then
    error "Installation failed"
    exit 1
fi