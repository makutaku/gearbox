#!/bin/bash

echo "Testing TUI Configuration Persistence"
echo "====================================="
echo ""

echo "Current MAX_PARALLEL_JOBS setting:"
grep "MAX_PARALLEL_JOBS" ~/.gearboxrc

echo ""
echo "📋 Instructions for manual testing:"
echo "1. Launch TUI: ./build/gearbox tui --demo"
echo "2. Navigate to Configuration view (C)"
echo "3. Find MAX_PARALLEL_JOBS setting - it should show '3' (current value)"
echo "4. Edit it to a different value (e.g., '6') using Enter"
echo "5. Save configuration with 's'"
echo "6. Quit TUI with 'q'"
echo "7. Run this script again to verify the change was saved"
echo ""
echo "Expected: MAX_PARALLEL_JOBS should show the new value in ~/.gearboxrc"
echo "Expected: When launching TUI again, it should load the saved value"
echo ""

if command -v tput >/dev/null 2>&1; then
    echo "📊 Launch TUI with: $(tput bold)./build/gearbox tui --demo$(tput sgr0)"
else
    echo "📊 Launch TUI with: ./build/gearbox tui --demo"
fi