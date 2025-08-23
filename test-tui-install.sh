#!/bin/bash

# Test TUI installation process to ensure clean display
echo "Testing TUI installation with output isolation..."

# First, make sure we have a tool that's not already installed
# For testing, let's uninstall a simple tool if it exists
echo "Preparing test environment..."

# Launch TUI in demo mode for safe testing without real installations
echo "Starting TUI in demo mode..."
echo "This should show a clean TUI interface without output corruption."
echo ""
echo "Instructions:"
echo "1. Navigate to Tool Browser (T)"
echo "2. Select a tool (Space)"
echo "3. Install it (i)"  
echo "4. Go to Monitor view (M) to watch progress"
echo "5. Verify the interface remains clean during 'installation'"
echo "6. Press 'q' to quit when done"
echo ""
echo "Press Enter to launch TUI in demo mode..."
read

# Launch in demo mode
./build/gearbox tui --demo