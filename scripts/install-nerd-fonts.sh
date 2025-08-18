#!/bin/bash

# Nerd Fonts Installation Script for Debian Linux
# Patched fonts with icons and glyphs for developers
# Usage: ./install-nerd-fonts.sh [OPTIONS]

set -e  # Exit on any error

# Find the script directory and load common library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

# Source common library for shared functions
if [[ -f "$REPO_DIR/lib/common.sh" ]]; then
    source "$REPO_DIR/lib/common.sh"
else
    echo "ERROR: common.sh not found in $REPO_DIR/lib/" >&2
    exit 1
fi

# Configuration
NERD_FONTS_REPO="https://github.com/ryanoasis/nerd-fonts"
NERD_FONTS_VERSION="v3.1.1"
FONTS_DIR="$HOME/.local/share/fonts"
DOWNLOAD_DIR="$HOME/tools/build/nerd-fonts"

# Font collections for different build types
declare -A MINIMAL_FONTS=(
    ["FiraCode"]="FiraCode Nerd Font - Programming ligatures + icons"
    ["JetBrainsMono"]="JetBrains Mono Nerd Font - Clean, readable monospace"
    ["Hack"]="Hack Nerd Font - Terminal-optimized"
)

declare -A STANDARD_FONTS=(
    ["FiraCode"]="FiraCode Nerd Font - Programming ligatures + icons"
    ["JetBrainsMono"]="JetBrains Mono Nerd Font - Clean, readable monospace"
    ["Hack"]="Hack Nerd Font - Terminal-optimized"
    ["SourceCodePro"]="Source Code Pro Nerd Font - Adobe's programming font"
    ["Inconsolata"]="Inconsolata Nerd Font - Classic monospace"
    ["CascadiaCode"]="Cascadia Code Nerd Font - Microsoft's new font"
    ["UbuntuMono"]="Ubuntu Mono Nerd Font - Ubuntu's monospace font"
    ["DejaVuSansMono"]="DejaVu Sans Mono Nerd Font - Popular open source font"
)

declare -A MAXIMUM_FONTS=(
    ["FiraCode"]="FiraCode Nerd Font - Programming ligatures + icons"
    ["JetBrainsMono"]="JetBrains Mono Nerd Font - Clean, readable monospace"
    ["Hack"]="Hack Nerd Font - Terminal-optimized"
    ["SourceCodePro"]="Source Code Pro Nerd Font - Adobe's programming font"
    ["Inconsolata"]="Inconsolata Nerd Font - Classic monospace"
    ["CascadiaCode"]="Cascadia Code Nerd Font - Microsoft's new font"
    ["UbuntuMono"]="Ubuntu Mono Nerd Font - Ubuntu's monospace font"
    ["DejaVuSansMono"]="DejaVu Sans Mono Nerd Font - Popular open source font"
    ["VictorMono"]="Victor Mono Nerd Font - Cursive italic programming font"
    ["Menlo"]="Menlo Nerd Font - Apple's monospace font"
    ["AnonymousPro"]="Anonymous Pro Nerd Font - Fixed-width font for coders"
    ["SpaceMono"]="Space Mono Nerd Font - Google's monospace font"
    ["IBMPlexMono"]="IBM Plex Mono Nerd Font - IBM's corporate font"
    ["RobotoMono"]="Roboto Mono Nerd Font - Google's robot-themed font"
    ["Terminus"]="Terminus Nerd Font - Bitmap font for coding"
)

# Default options
BUILD_TYPE="standard"
MODE="install"         # config, build, install
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
INTERACTIVE=false
SPECIFIC_FONTS=""
CONFIGURE_APPS=false
PREVIEW=false

# Show help
show_help() {
    cat << EOF
Patched fonts with icons and glyphs for developers

Usage: $0 [OPTIONS]

Build Types:
  --minimal            Essential fonts (3 fonts: FiraCode, JetBrains Mono, Hack)
  --standard           Popular fonts (8 fonts, ~80MB) - default
  --maximum            All fonts (15+ fonts, ~200MB)

Modes:
  -c, --config-only    Configure only (prepare build)
  -b, --build-only     Configure and build (no install)
  -i, --install        Configure, build, and install (default)

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
  $0 --maximum                     # Complete collection (15+ fonts)
  $0 --fonts="FiraCode,Hack"       # Install specific fonts
  $0 --interactive                 # Choose fonts interactively
  $0 --configure-apps              # Install and configure applications
  $0 --preview                     # Show font previews

Available Fonts:
$(printf "  %-20s %s\\n" "FiraCode" "Programming ligatures + icons")
$(printf "  %-20s %s\\n" "JetBrainsMono" "Clean, readable monospace")
$(printf "  %-20s %s\\n" "Hack" "Terminal-optimized")
$(printf "  %-20s %s\\n" "SourceCodePro" "Adobe's programming font")
$(printf "  %-20s %s\\n" "Inconsolata" "Classic monospace")
$(printf "  %-20s %s\\n" "CascadiaCode" "Microsoft's new font")
$(printf "  %-20s %s\\n" "UbuntuMono" "Ubuntu's monospace font")
$(printf "  %-20s %s\\n" "DejaVuSansMono" "Popular open source font")

EOF
}

# Parse command line arguments
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

# Prevent running as root
ensure_not_root

# Check for required commands
check_command() {
    if ! command -v "$1" &> /dev/null; then
        error "$1 is not installed. Please install it first."
    fi
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

    # Update package list
    log "Updating package list..."
    sudo apt update

    # Install required packages
    local packages=("fontconfig" "curl" "unzip" "fc-cache")
    local missing_packages=()
    
    for package in "${packages[@]}"; do
        if ! dpkg -l "$package" &>/dev/null; then
            missing_packages+=("$package")
        fi
    done
    
    if [[ ${#missing_packages[@]} -gt 0 ]]; then
        log "Installing missing packages: ${missing_packages[*]}"
        sudo apt install -y "${missing_packages[@]}"
        success "Dependencies installed successfully"
    else
        success "All dependencies are already installed"
    fi
}

# Get font collection based on build type
get_font_collection() {
    local build_type="$1"
    local -n fonts_ref=$2
    
    case "$build_type" in
        "minimal")
            for font in "${!MINIMAL_FONTS[@]}"; do
                fonts_ref["$font"]="${MINIMAL_FONTS[$font]}"
            done
            ;;
        "standard")
            for font in "${!STANDARD_FONTS[@]}"; do
                fonts_ref["$font"]="${STANDARD_FONTS[$font]}"
            done
            ;;
        "maximum")
            for font in "${!MAXIMUM_FONTS[@]}"; do
                fonts_ref["$font"]="${MAXIMUM_FONTS[$font]}"
            done
            ;;
        *)
            error "Unknown build type: $build_type"
            ;;
    esac
}

# Parse specific fonts
parse_specific_fonts() {
    local input="$1"
    local -n result_ref=$2
    
    IFS=',' read -ra font_array <<< "$input"
    for font in "${font_array[@]}"; do
        font=$(echo "$font" | xargs)  # Trim whitespace
        
        # Check if font exists in maximum collection
        if [[ -n "${MAXIMUM_FONTS[$font]:-}" ]]; then
            result_ref["$font"]="${MAXIMUM_FONTS[$font]}"
        else
            warning "Unknown font: $font. Available fonts: ${!MAXIMUM_FONTS[*]}"
        fi
    done
}

# Interactive font selection
interactive_font_selection() {
    local -n selected_fonts_ref=$1
    
    echo
    log "🎨 Interactive Font Selection"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo "Select fonts to install (press ENTER to toggle, 'q' to quit selection):"
    echo
    
    local -A font_selected
    local fonts_array=()
    local descriptions_array=()
    
    # Populate arrays with all available fonts
    for font in "${!MAXIMUM_FONTS[@]}"; do
        fonts_array+=("$font")
        descriptions_array+=("${MAXIMUM_FONTS[$font]}")
        font_selected["$font"]=false
    done
    
    # Pre-select standard fonts
    for font in "${!STANDARD_FONTS[@]}"; do
        font_selected["$font"]=true
    done
    
    local current_index=0
    local total_fonts=${#fonts_array[@]}
    
    while true; do
        # Clear screen and show header
        clear
        echo "🎨 Interactive Font Selection"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo
        echo "Use ↑/↓ to navigate, SPACE to toggle, ENTER to confirm, 'q' to quit"
        echo
        
        # Calculate selected size
        local selected_count=0
        local estimated_size=0
        for font in "${!font_selected[@]}"; do
            if [[ "${font_selected[$font]}" == true ]]; then
                ((selected_count++))
                ((estimated_size += 10))  # Rough estimate: 10MB per font
            fi
        done
        
        echo "Selected: $selected_count fonts (~${estimated_size}MB)"
        echo
        
        # Display font list
        for ((i=0; i<total_fonts; i++)); do
            local font="${fonts_array[$i]}"
            local description="${descriptions_array[$i]}"
            local prefix="   "
            
            if [[ $i -eq $current_index ]]; then
                prefix="→ "
            fi
            
            if [[ "${font_selected[$font]}" == true ]]; then
                echo -e "${prefix}${GREEN}✓${NC} ${BOLD}$font${NC} - $description"
            else
                echo -e "${prefix}◯ $font - $description"
            fi
        done
        
        echo
        echo "Controls: [↑/↓] Navigate  [SPACE] Toggle  [ENTER] Confirm  [q] Quit"
        
        # Read user input
        read -rsn1 key
        case "$key" in
            $'\x1b')  # ESC sequence
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
                local font="${fonts_array[$current_index]}"
                if [[ "${font_selected[$font]}" == true ]]; then
                    font_selected["$font"]=false
                else
                    font_selected["$font"]=true
                fi
                ;;
            '')  # Enter - confirm selection
                break
                ;;
            'q'|'Q')  # Quit
                echo
                warning "Font selection cancelled"
                exit 0
                ;;
        esac
    done
    
    # Copy selected fonts to result
    for font in "${!font_selected[@]}"; do
        if [[ "${font_selected[$font]}" == true ]]; then
            selected_fonts_ref["$font"]="${MAXIMUM_FONTS[$font]}"
        fi
    done
    
    clear
    echo "🎨 Font Selection Complete"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    success "Selected ${#selected_fonts_ref[@]} fonts for installation"
}

# Show font previews
show_font_previews() {
    local -A fonts_to_preview=("$@")
    
    echo
    log "🎨 Font Previews"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    for font in "${!fonts_to_preview[@]}"; do
        echo "${BOLD}$font Nerd Font:${NC}"
        echo "    → function() { return \"hello\"; } != <= >= => -> && ||"
        echo "    →     󰅲  󰘳  󰊕  󰗀  󰅴  󰆧  "
        echo
    done
    
    echo "Note: Actual appearance depends on terminal font support"
    echo
    read -p "Continue with installation? [Y/n] " -r
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        warning "Installation cancelled"
        exit 0
    fi
}

# Download and install fonts
download_and_install_fonts() {
    local -A fonts_to_install=("$@")
    local total_fonts=${#fonts_to_install[@]}
    local current_font=0
    
    log "🎨 Installing Nerd Fonts ($total_fonts fonts)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Create fonts directory
    mkdir -p "$FONTS_DIR"
    mkdir -p "$DOWNLOAD_DIR"
    
    for font in "${!fonts_to_install[@]}"; do
        ((current_font++))
        
        log "📦 [$current_font/$total_fonts] Installing $font..."
        
        # Check if font is already installed (unless force)
        if [[ "$FORCE_INSTALL" != true ]] && fc-list | grep -qi "$font.*nerd"; then
            success "  $font is already installed (use --force to reinstall)"
            continue
        fi
        
        # Download font
        local font_url="https://github.com/ryanoasis/nerd-fonts/releases/download/$NERD_FONTS_VERSION/$font.zip"
        local font_zip="$DOWNLOAD_DIR/$font.zip"
        
        log "  Downloading $font..."
        if ! curl -fsSL "$font_url" -o "$font_zip"; then
            error "Failed to download $font from $font_url"
        fi
        
        # Extract font
        local font_extract_dir="$DOWNLOAD_DIR/$font"
        rm -rf "$font_extract_dir"
        mkdir -p "$font_extract_dir"
        
        log "  Extracting $font..."
        if ! unzip -q "$font_zip" -d "$font_extract_dir"; then
            error "Failed to extract $font"
        fi
        
        # Install font files
        find "$font_extract_dir" -name "*.ttf" -o -name "*.otf" | while read -r font_file; do
            cp "$font_file" "$FONTS_DIR/"
        done
        
        # Clean up
        rm -f "$font_zip"
        rm -rf "$font_extract_dir"
        
        success "  $font installed successfully"
    done
    
    # Refresh font cache
    log "🔄 Refreshing font cache..."
    if fc-cache -fv > /dev/null 2>&1; then
        success "Font cache updated successfully"
    else
        warning "Font cache update failed, but fonts may still work"
    fi
}

# Verify installation
verify_installation() {
    local -A fonts_to_verify=("$@")
    local verified_count=0
    local total_fonts=${#fonts_to_verify[@]}
    
    log "🔍 Verifying font installation..."
    
    for font in "${!fonts_to_verify[@]}"; do
        if fc-list | grep -qi "$font.*nerd"; then
            success "  ✓ $font verified"
            ((verified_count++))
        else
            error "  ✗ $font not found"
        fi
    done
    
    if [[ $verified_count -eq $total_fonts ]]; then
        success "All $total_fonts fonts verified successfully!"
    else
        warning "$verified_count/$total_fonts fonts verified"
    fi
    
    # Show installation summary
    echo
    log "📊 Installation Summary"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "✅ %-20s %d fonts\n" "Installed:" $verified_count
    printf "💾 %-20s %s\n" "Location:" "$FONTS_DIR"
    printf "📏 %-20s %s\n" "Disk usage:" "$(du -sh "$FONTS_DIR" 2>/dev/null | cut -f1 || echo "Unknown")"
    echo
    
    # Show next steps
    echo "💡 Next Steps:"
    echo "  • Restart your terminal to see new fonts"
    echo "  • Configure VS Code: \"editor.fontFamily\": \"FiraCode Nerd Font\""
    echo "  • Configure terminal font to any installed Nerd Font"
    
    if command -v code &>/dev/null && [[ "$CONFIGURE_APPS" != true ]]; then
        echo "  • Run with --configure-apps to auto-configure applications"
    fi
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
            # Update existing settings
            if ! grep -q "editor.fontFamily" "$vscode_settings"; then
                # Add font family setting
                local temp_file=$(mktemp)
                jq '. + {"editor.fontFamily": "FiraCode Nerd Font", "editor.fontLigatures": true}' "$vscode_settings" > "$temp_file"
                mv "$temp_file" "$vscode_settings"
                success "  VS Code configuration updated"
            else
                warning "  VS Code font already configured"
            fi
        else
            # Create new settings file
            cat > "$vscode_settings" << EOF
{
    "editor.fontFamily": "FiraCode Nerd Font",
    "editor.fontLigatures": true,
    "terminal.integrated.fontFamily": "JetBrains Mono Nerd Font"
}
EOF
            success "  VS Code configuration created"
        fi
    fi
    
    # Terminal configurations
    configure_terminal_fonts
}

# Configure terminal fonts
configure_terminal_fonts() {
    # Kitty terminal
    if command -v kitty &>/dev/null; then
        local kitty_config="$HOME/.config/kitty/kitty.conf"
        local kitty_dir="$(dirname "$kitty_config")"
        
        mkdir -p "$kitty_dir"
        
        if [[ -f "$kitty_config" ]]; then
            if ! grep -q "font_family.*Nerd" "$kitty_config"; then
                echo "font_family JetBrains Mono Nerd Font" >> "$kitty_config"
                success "  Kitty terminal font configured"
            fi
        else
            echo "font_family JetBrains Mono Nerd Font" > "$kitty_config"
            success "  Kitty configuration created"
        fi
    fi
    
    # Alacritty terminal
    if command -v alacritty &>/dev/null; then
        local alacritty_config="$HOME/.config/alacritty/alacritty.yml"
        local alacritty_dir="$(dirname "$alacritty_config")"
        
        mkdir -p "$alacritty_dir"
        
        if [[ ! -f "$alacritty_config" ]] || ! grep -q "family.*Nerd" "$alacritty_config"; then
            cat >> "$alacritty_config" << EOF

# Nerd Font configuration
font:
  normal:
    family: JetBrains Mono Nerd Font
  bold:
    family: JetBrains Mono Nerd Font
  italic:
    family: JetBrains Mono Nerd Font
EOF
            success "  Alacritty terminal font configured"
        fi
    fi
}

# Main execution
main() {
    log "🎨 Nerd Fonts Installation"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Detect system configuration
    log "📋 Detecting system configuration..."
    log "  OS: $(lsb_release -d 2>/dev/null | cut -f2 || echo "Unknown Linux")"
    log "  Font directory: $FONTS_DIR"
    log "  Available space: $(df -h "$HOME" | awk 'NR==2 {print $4}' || echo "Unknown")"
    echo
    
    # Install dependencies
    if [[ "$MODE" != "config" ]]; then
        install_dependencies
    fi
    
    # Determine fonts to install
    declare -A fonts_to_install
    
    if [[ -n "$SPECIFIC_FONTS" ]]; then
        parse_specific_fonts "$SPECIFIC_FONTS" fonts_to_install
        log "Installing specific fonts: ${!fonts_to_install[*]}"
    elif [[ "$INTERACTIVE" == true ]]; then
        interactive_font_selection fonts_to_install
    else
        get_font_collection "$BUILD_TYPE" fonts_to_install
        log "Installing $BUILD_TYPE collection: ${!fonts_to_install[*]}"
    fi
    
    if [[ ${#fonts_to_install[@]} -eq 0 ]]; then
        error "No fonts selected for installation"
    fi
    
    # Show preview if requested
    if [[ "$PREVIEW" == true ]]; then
        show_font_previews "${fonts_to_install[@]}"
    fi
    
    # Exit early if config-only mode
    if [[ "$MODE" == "config" ]]; then
        success "Configuration completed. Run without --config-only to install fonts."
        exit 0
    fi
    
    # Download and install fonts
    if [[ "$MODE" != "build" ]]; then
        download_and_install_fonts "${fonts_to_install[@]}"
        
        # Verify installation
        if [[ "$RUN_TESTS" == true ]]; then
            verify_installation "${fonts_to_install[@]}"
        fi
        
        # Configure applications
        if [[ "$CONFIGURE_APPS" == true ]]; then
            configure_applications
        fi
        
        success "🎉 Nerd Fonts installation completed!"
    else
        success "Build preparation completed. Run without --build-only to install fonts."
    fi
}

# Run main function
main "$@"