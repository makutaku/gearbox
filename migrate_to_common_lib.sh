#!/bin/bash

# Script to migrate all installation scripts to use lib/common.sh
# This eliminates code duplication and standardizes logging

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Function to migrate a single script
migrate_script() {
    local script_file="$1"
    local script_name=$(basename "$script_file")
    
    echo "Migrating $script_name..."
    
    # Skip if already migrated
    if grep -q "lib/common.sh" "$script_file"; then
        echo "  Already migrated: $script_name"
        return 0
    fi
    
    # Skip special scripts
    case "$script_name" in
        "install-fd.sh"|"install-ripgrep.sh"|"install-all-tools.sh")
            echo "  Already manually migrated: $script_name"
            return 0
            ;;
    esac
    
    # Create backup
    cp "$script_file" "$script_file.backup"
    
    # Read the entire file
    local temp_file=$(mktemp)
    
    # Replace the color definitions and add common library source
    awk '
    BEGIN { 
        in_color_section = 0
        in_log_section = 0
        color_section_done = 0
        common_lib_added = 0
        skip_until_blank = 0
    }
    
    # Handle set -e line - add common library after it
    /^set -e/ {
        print $0
        if (!common_lib_added) {
            print ""
            print "# Find the script directory and load common library"
            print "SCRIPT_DIR=\"$(cd \"$(dirname \"${BASH_SOURCE[0]}\")\" && pwd)\""
            print "REPO_DIR=\"$(dirname \"$SCRIPT_DIR\")\""
            print ""
            print "# Source common library for shared functions"
            print "if [[ -f \"$REPO_DIR/lib/common.sh\" ]]; then"
            print "    source \"$REPO_DIR/lib/common.sh\""
            print "else"
            print "    echo \"ERROR: common.sh not found in $REPO_DIR/lib/\" >&2"
            print "    exit 1"
            print "fi"
            common_lib_added = 1
        }
        next
    }
    
    # Skip color definitions
    /^# Colors for output/ { in_color_section = 1; next }
    /^RED=|^GREEN=|^YELLOW=|^BLUE=|^NC=/ && in_color_section { next }
    
    # Skip logging functions
    /^# Logging function/ { in_log_section = 1; skip_until_blank = 1; next }
    /^log\(\)|^error\(\)|^success\(\)|^warning\(\)/ { in_log_section = 1; skip_until_blank = 1; next }
    
    # Skip function bodies and blank lines in log section
    in_log_section && skip_until_blank {
        if (/^[[:space:]]*$/ && !in_function) {
            in_log_section = 0
            skip_until_blank = 0
            print "# Note: Logging functions now provided by lib/common.sh"
            print ""
            next
        }
        if (/^[[:space:]]*}[[:space:]]*$/) {
            in_function = 0
        } else if (/{/) {
            in_function = 1
        }
        next
    }
    
    # Reset color section flag after colors
    in_color_section && (/^[[:space:]]*$/ || !/^(RED|GREEN|YELLOW|BLUE|NC)=/) {
        in_color_section = 0
        if (!color_section_done) {
            color_section_done = 1
        }
    }
    
    # Print everything else
    !in_color_section && !in_log_section {
        print $0
    }
    ' "$script_file" > "$temp_file"
    
    # Move the temporary file back
    mv "$temp_file" "$script_file"
    
    echo "  Successfully migrated: $script_name"
}

# Migrate all installation scripts
for script in scripts/install-*.sh; do
    if [[ -f "$script" ]]; then
        migrate_script "$script"
    fi
done

echo "Migration complete!"
echo "All scripts now use lib/common.sh for shared functionality."