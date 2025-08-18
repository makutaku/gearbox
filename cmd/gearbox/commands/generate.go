package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewGenerateCmd creates the generate command
func NewGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [TOOLS...]",
		Short: "Generate optimized installation scripts from templates",
		Long: `Generate installation scripts for tools using the template system.

This command uses the script-generator to create optimized installation
scripts from templates, supporting multiple languages and build systems.`,
		RunE: runGenerate,
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be generated without creating files")
	cmd.Flags().Bool("force", false, "Overwrite existing scripts")
	cmd.Flags().Bool("validate", true, "Validate generated scripts")
	cmd.Flags().StringP("output", "o", "", "Output directory (default: scripts/)")
	cmd.Flags().StringP("template-dir", "t", "", "Template directory (default: templates/)")

	return cmd
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Get the directory where the gearbox binary is located
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	generatorPath := filepath.Join(repoDir, "bin", "script-generator")

	// This command requires the script-generator
	if _, err := os.Stat(generatorPath); err != nil {
		return fmt.Errorf(`script generation requires the script-generator tool

The script-generator is not available. To enable script generation:
1. Ensure Go is installed 
2. Run: cd tools/script-generator && go build -o ../../bin/script-generator
3. Or rebuild the entire project

For more information, run: gearbox doctor`)
	}

	return runWithGenerator(generatorPath, cmd, args)
}

func runWithGenerator(generatorPath string, cmd *cobra.Command, args []string) error {
	genCmd := exec.Command(generatorPath, "generate")
	
	// Add tool arguments
	genCmd.Args = append(genCmd.Args, args...)

	// Pass through flags
	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		genCmd.Args = append(genCmd.Args, "--dry-run")
	}
	if force, _ := cmd.Flags().GetBool("force"); force {
		genCmd.Args = append(genCmd.Args, "--force")
	}
	if validate, _ := cmd.Flags().GetBool("validate"); !validate {
		genCmd.Args = append(genCmd.Args, "--validate=false")
	}
	if output, _ := cmd.Flags().GetString("output"); output != "" {
		genCmd.Args = append(genCmd.Args, "--output", output)
	}
	if templateDir, _ := cmd.Flags().GetString("template-dir"); templateDir != "" {
		genCmd.Args = append(genCmd.Args, "--template-dir", templateDir)
	}

	genCmd.Stdout = os.Stdout
	genCmd.Stderr = os.Stderr
	
	return genCmd.Run()
}