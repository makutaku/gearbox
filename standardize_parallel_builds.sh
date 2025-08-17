#!/bin/bash

# Script to standardize parallel build usage across all installation scripts
# Replace manual nproc usage with get_optimal_jobs function

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Function to standardize parallel builds in a script
standardize_parallel_builds() {
    local script_file="$1"
    local script_name=$(basename "$script_file")
    
    echo "Standardizing parallel builds in $script_name..."
    
    # Skip if script doesn't use parallel builds
    if ! grep -q "nproc\|NPROC\|-j" "$script_file"; then
        echo "  No parallel builds found in $script_name"
        return 0
    fi
    
    # Create backup
    cp "$script_file" "$script_file.parallel_backup"
    
    # Replace NPROC=$(nproc) with standard pattern
    sed -i 's/NPROC=$(nproc)/CORES=$(get_optimal_jobs)/' "$script_file"
    sed -i 's/nproc)/get_optimal_jobs)/' "$script_file"
    sed -i 's/\$NPROC/\$CORES/g' "$script_file"
    
    # Update log messages
    sed -i 's/Using \$NPROC CPU cores/Using \$CORES CPU cores/' "$script_file"
    
    echo "  Successfully standardized: $script_name"
}

# Standardize all installation scripts
for script in scripts/install-*.sh; do
    if [[ -f "$script" ]]; then
        standardize_parallel_builds "$script"
    fi
done

echo "Parallel build standardization complete!"
echo "All scripts now use get_optimal_jobs() for consistent parallel builds."