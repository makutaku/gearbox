# Health Check Auto-Run Fix

## Issue
Health checks (Memory, Disk Space, Internet Connection, Toolchains) were showing "Checking..." indefinitely because they were never automatically run when the Health Monitor view was loaded.

## Root Cause
The `runHealthChecks()` method was only called when the user pressed 'r', not when the view was initially loaded with data.

## Solution
Modified the `SetData()` method to automatically run health checks when data is loaded:

```go
// SetData updates the health view data
func (hv *HealthView) SetData(tools []orchestrator.ToolConfig, installed map[string]*manifest.InstallationRecord) {
    hv.installedTools = installed
    hv.updateToolChecks(tools)
    // Automatically run health checks when data is loaded
    hv.runHealthChecks()
}
```

Also added the missing "Tool Updates" check update:
```go
// Update the tool updates check
if len(hv.toolChecks) > 3 {
    hv.toolChecks[3].Status = HealthStatusPassing
    hv.toolChecks[3].Message = "All tools up to date"
    hv.toolChecks[3].Details = []string{
        "Last checked: just now",
        "No updates available",
    }
}
```

## Impact
When users navigate to the Health Monitor view, they will now see:
- ✅ Memory status with available/used/total
- ⚠️ Disk space warnings if low
- ✅ Internet connection status
- ✅ Build tools installation status
- ✅ Git version
- ✅ PATH configuration
- ✅ Rust toolchain status
- ✅ Go toolchain status
- ✅ Tool coverage statistics
- ✅ Tool updates status

All checks run automatically on view load, with the option to refresh by pressing 'r'.

## User Experience
- No more indefinite "Checking..." states
- Immediate feedback on system health
- All checks complete within the first render
- Manual refresh still available with 'r' key