# Tool-Specific Doctor Framework

## Overview

The Gearbox doctor command supports comprehensive health checks for individual tools. This document explains how to implement tool-specific diagnostics.

## Current Implementation

### Supported Tools
- **nerd-fonts**: Font installation, cache status, and availability checks
- **zoxide**: Database status, shell integration, and performance checks

### Usage
```bash
# General health check
gearbox doctor

# Tool-specific diagnostics  
gearbox doctor zoxide
gearbox doctor nerd-fonts

# Verbose mode with detailed information
gearbox doctor zoxide --verbose
gearbox doctor nerd-fonts --verbose

# Auto-fix mode (where supported)
gearbox doctor zoxide --fix
```

## Adding New Tool Diagnostics

### 1. Update the Switch Statement
In `cmd/gearbox/commands/doctor.go`, add your tool to the `runToolSpecificDoctor` function:

```go
func runToolSpecificDoctor(repoDir, toolName string, cmd *cobra.Command) error {
    switch toolName {
    case "nerd-fonts":
        return runNerdFontsDoctor(repoDir, cmd)
    case "zoxide":
        return runZoxideDoctor(cmd)
    case "your-tool":  // Add your tool here
        return runYourToolDoctor(cmd)
    default:
        return fmt.Errorf("tool-specific diagnostics not implemented for '%s'", toolName)
    }
}
```

### 2. Implement the Doctor Function

Create a comprehensive doctor function following this template:

```go
// runYourToolDoctor performs comprehensive health checks for your-tool
func runYourToolDoctor(cmd *cobra.Command) error {
    verbose, _ := cmd.Flags().GetBool("verbose")
    fix, _ := cmd.Flags().GetBool("fix")
    
    fmt.Println("🔍 Your Tool Health Check")
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    fmt.Println()
    
    // Track overall health
    var totalChecks, passedChecks, failedChecks, warningChecks int
    var issues []string
    var suggestions []string
    
    // 1. Installation Status Check
    fmt.Println("📍 Installation Status")
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    
    totalChecks++
    toolPath, err := exec.LookPath("your-tool")
    if err != nil {
        fmt.Printf("❌ Your Tool not found in PATH\n")
        failedChecks++
        issues = append(issues, "your-tool binary not found")
        suggestions = append(suggestions, "Install your-tool: ./build/gearbox install your-tool")
    } else {
        fmt.Printf("✅ Your Tool found at: %s\n", toolPath)
        passedChecks++
        
        // Check version
        totalChecks++
        version, err := exec.Command("your-tool", "--version").Output()
        if err != nil {
            fmt.Printf("⚠️  Could not get your-tool version: %v\n", err)
            warningChecks++
        } else {
            versionStr := strings.TrimSpace(string(version))
            fmt.Printf("✅ Version: %s\n", versionStr)
            passedChecks++
        }
    }
    fmt.Println()
    
    // 2. Add more specific checks for your tool
    // - Configuration files
    // - Dependencies
    // - Data files
    // - Integration status
    // - Performance checks
    
    // Summary
    fmt.Println("📊 Health Summary")
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    fmt.Printf("Total Checks: %d\n", totalChecks)
    fmt.Printf("✅ Passed: %d\n", passedChecks)
    fmt.Printf("⚠️  Warnings: %d\n", warningChecks)
    fmt.Printf("❌ Failed: %d\n", failedChecks)
    
    if len(issues) > 0 {
        fmt.Println("\n🔧 Issues Detected:")
        for _, issue := range issues {
            fmt.Printf("  • %s\n", issue)
        }
    }
    
    if len(suggestions) > 0 {
        fmt.Println("\n💡 Suggestions:")
        for _, suggestion := range suggestions {
            fmt.Printf("  • %s\n", suggestion)
        }
    }
    
    if fix && len(issues) > 0 {
        fmt.Println("\n🔧 Auto-fix functionality")
        // Implement auto-fix logic here
    }
    
    // Return error if critical issues found
    if failedChecks > 0 {
        return fmt.Errorf("your-tool health check failed with %d critical issues", failedChecks)
    }
    
    fmt.Println("\n🎉 Your Tool health check completed!")
    return nil
}
```

### 3. Update Help Documentation

Add your tool to the help text in the `NewDoctorCmd` function:

```go
Tool-specific diagnostics:
- nerd-fonts: Font installation, cache status, and availability checks
- zoxide: Database status, shell integration, and performance checks
- your-tool: Description of what your tool diagnostics check

Examples:
  gearbox doctor                    # General system health check
  gearbox doctor your-tool          # Your tool specific diagnostics
  gearbox doctor your-tool --verbose  # Detailed analysis
```

## Check Categories

### Common Check Types
1. **Installation Status**
   - Binary location and PATH availability
   - Version detection and validation
   - Installation method verification

2. **Configuration Checks**
   - Config file existence and validity
   - Settings verification
   - Permission checks

3. **Integration Checks**
   - Shell integration status
   - Plugin/extension compatibility
   - System service integration

4. **Data Integrity**
   - Database or cache status
   - Data file validation
   - Index integrity

5. **Performance Checks**
   - Response time validation
   - Resource usage analysis
   - Optimization suggestions

6. **Dependency Checks**
   - Required dependencies available
   - Version compatibility
   - Optional dependencies

## Best Practices

### Error Handling
- Use different severity levels: ✅ Passed, ⚠️ Warning, ❌ Failed
- Provide clear, actionable error messages
- Track statistics for comprehensive reporting

### User Experience
- Use clear section headers with Unicode boxes
- Provide verbose mode for detailed information
- Include helpful suggestions for fixing issues
- Show progress with check counts

### Auto-Fix Support
- Implement safe auto-fix operations where possible
- Always show what would be fixed before applying
- Provide fallback manual instructions

### Performance
- Make checks efficient and non-destructive
- Cache results where appropriate
- Provide quick vs comprehensive check options

## Example Implementation: Zoxide

The zoxide doctor implementation demonstrates:
- ✅ Installation detection with PATH checking
- ✅ Version verification
- ✅ Database status and entry counting
- ✅ Shell integration detection across multiple shells
- ✅ Alias functionality verification  
- ✅ Performance testing
- ✅ Verbose mode with database contents
- ✅ Comprehensive suggestions and issue tracking

See `runZoxideDoctor()` in `cmd/gearbox/commands/doctor.go` for the complete implementation.

## Testing Your Implementation

1. Build the project: `make build`
2. Test basic functionality: `./build/gearbox doctor your-tool`
3. Test verbose mode: `./build/gearbox doctor your-tool --verbose`
4. Test with tool not installed
5. Test with various configuration states
6. Verify help text shows correctly