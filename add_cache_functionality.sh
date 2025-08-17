#!/bin/bash

# Script to add build cache functionality to all installation scripts
# This script adds cache checking and storage to improve build performance

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPTS_DIR="$SCRIPT_DIR/scripts"

# Function to add cache functionality to a script
add_cache_to_script() {
    local script_file="$1"
    local tool_name="$2"
    local binary_names="$3"  # Space-separated list of binaries
    
    echo "Adding cache functionality to $script_file for tool: $tool_name"
    
    # Check if script already has cache functionality
    if grep -q "is_cached" "$script_file"; then
        echo "  - Cache functionality already present in $script_file"
        return 0
    fi
    
    # Find the installation section (look for patterns like "Installing tool" or "make install")
    local install_line=$(grep -n "log.*Installing\|sudo make install\|sudo cp.*bin" "$script_file" | head -1 | cut -d: -f1)
    
    if [[ -z "$install_line" ]]; then
        echo "  - Could not find installation section in $script_file"
        return 1
    fi
    
    echo "  - Found installation section at line $install_line"
    
    # Create backup
    cp "$script_file" "${script_file}.backup-cache"
    
    # Read the file into array
    mapfile -t lines < "$script_file"
    
    # Find installation commands and wrap with cache logic
    local modified=false
    local new_lines=()
    local i=0
    
    while [[ $i -lt ${#lines[@]} ]]; do
        local line="${lines[$i]}"
        
        # Check if this is an installation command we should wrap with cache
        if [[ "$line" =~ ^[[:space:]]*sudo[[:space:]]+(cp|make[[:space:]]+install) ]] && [[ ! "$line" =~ apt|ldconfig|chmod ]]; then
            if [[ ! "$modified" ]]; then
                # Add cache checking logic before first installation command
                new_lines+=("")
                new_lines+=("# Check if we can use cached binary")
                new_lines+=("if is_cached \"$tool_name\" \"\$BUILD_TYPE\"; then")
                new_lines+=("    log \"Found cached $tool_name build, using cached version...\"")
                
                # Build cache retrieval command
                local cache_cmd="    if "
                local first=true
                for binary in $binary_names; do
                    if [[ "$first" == true ]]; then
                        cache_cmd+="get_cached_binary \"$binary\" \"\$BUILD_TYPE\" \"/usr/local/bin/$binary\""
                        first=false
                    else
                        cache_cmd+=" && \\"
                        new_lines+=("$cache_cmd")
                        cache_cmd="       get_cached_binary \"$binary\" \"\$BUILD_TYPE\" \"/usr/local/bin/$binary\""
                    fi
                done
                cache_cmd+="; then"
                new_lines+=("$cache_cmd")
                
                # Add chmod for cached binaries
                local chmod_cmd="        sudo chmod +x"
                for binary in $binary_names; do
                    chmod_cmd+=" /usr/local/bin/$binary"
                done
                chmod_cmd+=" || error \"Failed to set executable permissions\""
                new_lines+=("$chmod_cmd")
                new_lines+=("        success \"$tool_name installed from cache successfully\"")
                new_lines+=("    else")
                new_lines+=("        warning \"Failed to use cached binary, proceeding with fresh installation\"")
                
                modified=true
            fi
            
            # Add the original installation command with indentation
            new_lines+=("        $line")
            
            # Look ahead for related installation commands
            local j=$((i + 1))
            while [[ $j -lt ${#lines[@]} ]]; do
                local next_line="${lines[$j]}"
                if [[ "$next_line" =~ ^[[:space:]]*sudo[[:space:]]+(cp|chmod) ]] && [[ ! "$next_line" =~ apt|ldconfig ]]; then
                    new_lines+=("        $next_line")
                    j=$((j + 1))
                else
                    break
                fi
            done
            
            # Add cache storage command
            local cache_store_cmd="        # Cache the new build"
            new_lines+=("$cache_store_cmd")
            cache_store_cmd="        cache_build \"$tool_name\" \"\$BUILD_TYPE\""
            for binary in $binary_names; do
                cache_store_cmd+=" \"/usr/local/bin/$binary\""
            done
            new_lines+=("$cache_store_cmd")
            new_lines+=("    fi")
            new_lines+=("else")
            
            # Add original commands for else block
            new_lines+=("    $line")
            local k=$((i + 1))
            while [[ $k -lt ${#lines[@]} ]]; do
                local else_line="${lines[$k]}"
                if [[ "$else_line" =~ ^[[:space:]]*sudo[[:space:]]+(cp|chmod) ]] && [[ ! "$else_line" =~ apt|ldconfig ]]; then
                    new_lines+=("    $else_line")
                    k=$((k + 1))
                else
                    break
                fi
            done
            
            # Add cache storage for else block
            new_lines+=("    # Cache the new build")
            cache_store_cmd="    cache_build \"$tool_name\" \"\$BUILD_TYPE\""
            for binary in $binary_names; do
                cache_store_cmd+=" \"/usr/local/bin/$binary\""
            done
            new_lines+=("$cache_store_cmd")
            new_lines+=("fi")
            
            # Skip the lines we've already processed
            i=$((j - 1))
        else
            new_lines+=("$line")
        fi
        
        i=$((i + 1))
    done
    
    # Write the modified content back to file
    printf '%s\n' "${new_lines[@]}" > "$script_file"
    
    echo "  - Cache functionality added successfully"
}

# Tool configurations: script_name:tool_name:binary_names
declare -A TOOL_CONFIGS=(
    ["install-fd.sh"]="fd:fd"
    ["install-ripgrep.sh"]="ripgrep:rg"
    ["install-fzf.sh"]="fzf:fzf"
    ["install-bat.sh"]="bat:bat"
    ["install-eza.sh"]="eza:eza"
    ["install-delta.sh"]="delta:delta"
    ["install-starship.sh"]="starship:starship"
    ["install-zoxide.sh"]="zoxide:zoxide"
    ["install-bottom.sh"]="bottom:btm"
    ["install-procs.sh"]="procs:procs"
    ["install-tokei.sh"]="tokei:tokei"
    ["install-hyperfine.sh"]="hyperfine:hyperfine"
    ["install-dust.sh"]="dust:dust"
    ["install-sd.sh"]="sd:sd"
    ["install-tealdeer.sh"]="tealdeer:tldr"
    ["install-choose.sh"]="choose:choose"
    ["install-difftastic.sh"]="difftastic:difft"
    ["install-bandwhich.sh"]="bandwhich:bandwhich"
    ["install-xsv.sh"]="xsv:xsv"
    ["install-gh.sh"]="gh:gh"
    ["install-lazygit.sh"]="lazygit:lazygit"
    ["install-ruff.sh"]="ruff:ruff"
    ["install-uv.sh"]="uv:uv"
    ["install-fclones.sh"]="fclones:fclones"
    ["install-serena.sh"]="serena:serena"
    ["install-ffmpeg.sh"]="ffmpeg:ffmpeg:ffprobe"
    ["install-7zip.sh"]="7zip:7zz"
    ["install-imagemagick.sh"]="imagemagick:magick:convert:identify"
)

echo "Adding cache functionality to installation scripts..."
echo

# Process each script
for script_config in "${!TOOL_CONFIGS[@]}"; do
    local script_file="$SCRIPTS_DIR/$script_config"
    
    if [[ ! -f "$script_file" ]]; then
        echo "Warning: Script not found: $script_file"
        continue
    fi
    
    # Parse configuration
    IFS=':' read -r tool_name binary_names <<< "${TOOL_CONFIGS[$script_config]}"
    
    add_cache_to_script "$script_file" "$tool_name" "$binary_names"
    echo
done

echo "Cache functionality addition completed!"
echo
echo "Note: Backup files created with .backup-cache extension"
echo "Review the changes and remove backup files if satisfied"