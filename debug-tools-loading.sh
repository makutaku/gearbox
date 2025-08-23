#!/bin/bash

echo "🔍 Debug Tools Loading Issue"
echo "============================="
echo ""

echo "Testing if the issue is fixed:"
echo ""

echo "1. Launch TUI and check tools page:"
echo "   ./build/gearbox tui"
echo "   Press 't' to go to tools page"
echo "   Should show loading message first, then tools should appear"
echo ""

echo "2. Check tool browser data flow:"
echo "   - SetData() should populate tb.tools and tb.installedTools" 
echo "   - LoadFullContent() should call applyFilters() to populate tb.filteredTools"
echo "   - updateViewportContentTUI() should render the filtered tools"
echo ""

echo "3. Expected behavior:"
echo "   ✅ 't' press shows loading message instantly"
echo "   ✅ After 1-2 seconds, tools list should appear"
echo "   ✅ Tools should be visible with status indicators (✓ or ○)"
echo "   ✅ Navigation with ↑/↓ should work"
echo ""

echo "4. If still broken, possible causes:"
echo "   • tb.tools is empty (SetData not called)"
echo "   • tb.filteredTools is empty (applyFilters issue)" 
echo "   • async command not executing (loadToolBrowserContentAsync)"
echo "   • viewport not updating (updateViewportContentTUI issue)"
echo ""

echo "🚀 Test now: ./build/gearbox tui → press 't'"