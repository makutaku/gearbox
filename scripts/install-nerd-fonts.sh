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

# Available fonts - simplified to arrays instead of associative arrays
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
)

readonly -a MAXIMUM_FONTS=(
    "FiraCode"
    "JetBrainsMono"
    "Hack"
    "SourceCodePro"
    "Inconsolata"
    "CascadiaCode"
    "UbuntuMono"
    "DejaVuSansMono"
    "VictorMono"
    "Menlo"
    "AnonymousPro"
    "SpaceMono"
    "IBMPlexMono"
    "RobotoMono"
    "Terminus"
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
  --standard           Popular fonts (8 fonts, ~80MB) - default
  --maximum            All fonts (15+ fonts, ~200MB)

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

# Interactive font selection
interactive_font_selection() {
    echo
    log "🎨 Interactive Font Selection"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo "Select fonts to install (press ENTER to toggle, 'q' to quit selection):"
    echo
    
    # Create arrays for fonts and their selection status
    local -a available_fonts=("${MAXIMUM_FONTS[@]}")
    local -a font_selected=()
    
    # Initialize selection status (pre-select standard fonts)
    for font in "${available_fonts[@]}"; do
        local is_standard=false
        for standard_font in "${STANDARD_FONTS[@]}"; do
            if [[ "$font" == "$standard_font" ]]; then
                is_standard=true
                break
            fi
        done
        font_selected+=("$is_standard")
    done
    
    local current_index=0
    local total_fonts=${#available_fonts[@]}
    
    while true; do
        # Clear screen and show header
        clear
        echo "🎨 Interactive Font Selection"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo
        echo "Use ↑/↓ to navigate, SPACE to toggle, 'p' to preview, ENTER to confirm, 'q' to quit"
        echo
        
        # Calculate selected count and size
        local selected_count=0
        local estimated_size=0
        for ((i=0; i<total_fonts; i++)); do
            if [[ "${font_selected[i]}" == true ]]; then
                ((selected_count++))
                ((estimated_size += 10))  # Rough estimate: 10MB per font
            fi
        done
        
        echo "Selected: $selected_count fonts (~${estimated_size}MB)"
        echo
        
        # Display font list
        for ((i=0; i<total_fonts; i++)); do
            local font="${available_fonts[i]}"
            local prefix="   "
            
            if [[ $i -eq $current_index ]]; then
                prefix="→ "
            fi
            
            if [[ "${font_selected[i]}" == true ]]; then
                echo -e "${prefix}✓ ${font}"
            else
                echo -e "${prefix}◯ ${font}"
            fi
        done
        
        echo
        echo "Controls: [↑/↓] Navigate  [SPACE] Toggle  [p] Preview  [ENTER] Confirm  [q] Quit"
        
        # Read user input
        read -rsn1 key
        case "$key" in
            $'\x1b')  # ESC sequence for arrow keys
                read -rsn2 -t 0.1 key
                case "$key" in
                    '[A')  # Up arrow
                        ((current_index > 0)) && ((current_index--))
                        ;;
                    '[B')  # Down arrow
                        ((current_index < total_fonts - 1)) && ((current_index++))
                        ;;
                esac
                ;;
            ' ')  # Space - toggle selection
                if [[ "${font_selected[current_index]}" == true ]]; then
                    font_selected[current_index]=false
                else
                    font_selected[current_index]=true
                fi
                ;;
            '')  # Enter - confirm selection
                break
                ;;
            'p'|'P')  # Preview current font
                local current_font="${available_fonts[current_index]}"
                echo
                log "🔍 Font Preview: $current_font"
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                
                echo "${current_font} Nerd Font:"
                echo "  Code: const fn = () => result !== null && value >= 0"
                echo "  Icons:   󰈙  󰗀  󰅴  󰘳  󰊕  󰄛  󰊢"
                echo "  Arrows: => -> <- ↗ ↙ ⟨ ⟩ ⮕"
                
                case "$current_font" in
                    "FiraCode")
                        echo "  Special: Programming ligatures and coding symbols"
                        echo "  Perfect for: Code editors, terminals, IDEs"
                        ;;
                    "JetBrainsMono")
                        echo "  Special: Clean lines, excellent readability"
                        echo "  Perfect for: Long coding sessions, professional use"
                        ;;
                    "Hack")
                        echo "  Special: Optimized for terminals"
                        echo "  Perfect for: Command line work, system administration"
                        ;;
                    *)
                        echo "  Description: Programming font with Nerd Font icons"
                        ;;
                esac
                
                echo
                read -p "Press ENTER to return to selection menu..." -r
                ;;
            'q'|'Q')  # Quit
                echo
                warning "Font selection cancelled"
                exit 0
                ;;
        esac
    done
    
    # Output selected fonts
    clear
    echo "🎨 Font Selection Complete"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    local selected_fonts=()
    for ((i=0; i<total_fonts; i++)); do
        if [[ "${font_selected[i]}" == true ]]; then
            selected_fonts+=("${available_fonts[i]}")
        fi
    done
    
    if [[ ${#selected_fonts[@]} -eq 0 ]]; then
        warning "No fonts selected"
        return 1
    fi
    
    success "Selected ${#selected_fonts[@]} fonts for installation"
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