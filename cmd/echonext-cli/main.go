package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

var version = "dev" // Will be set during build via ldflags

func getVersion() string {
	// If version was set via ldflags, use it
	if version != "dev" {
		return version
	}

	// Try to get version from build info (works with go install @version)
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}

	return version
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "echonext",
		Short: "EchoNext CLI - Build type-safe APIs with automatic OpenAPI generation",
		Long: `EchoNext CLI is a command-line tool for building production-ready APIs
with automatic OpenAPI generation, type-safe handlers, and domain-driven architecture.

Built on top of Echo and GORM with zero vendor lock-in.`,
	}

	// Add commands
	rootCmd.AddCommand(
		newInitCmd(),
		newGenerateCmd(),
		newVersionCmd(),
		newDBCmd(),
		newDevCmd(),
		newDocsCmd(),
		newTestCmd(),
		newBuildCmd(),
		newUpgradeCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// newVersionCmd returns the version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("EchoNext CLI %s\n", getVersion())
			fmt.Println("Build type-safe APIs with automatic OpenAPI generation")
			fmt.Println("https://github.com/abdussamadbello/echonext")
		},
	}
}

// newUpgradeCmd returns the upgrade command
func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade EchoNext CLI to the latest version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Upgrading EchoNext CLI...")
			fmt.Printf("Current version: %s\n\n", getVersion())

			// Get current executable path
			currentExe, err := os.Executable()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not determine current executable path: %v\n", err)
			}

			// Run go install to get latest version
			installCmd := exec.Command("go", "install", "github.com/abdussamadbello/echonext/cmd/echonext-cli@latest")
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr

			if err := installCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Upgrade failed: %v\n", err)
				os.Exit(1)
			}

			// Get GOBIN path
			goBinCmd := exec.Command("go", "env", "GOBIN")
			goBinOutput, _ := goBinCmd.Output()
			goBin := strings.TrimSpace(string(goBinOutput))
			if goBin == "" {
				goPathCmd := exec.Command("go", "env", "GOPATH")
				goPathOutput, _ := goPathCmd.Output()
				goBin = filepath.Join(strings.TrimSpace(string(goPathOutput)), "bin")
			}

			// Determine source binary name based on OS
			sourceBinary := filepath.Join(goBin, "echonext-cli")
			if runtime.GOOS == "windows" {
				sourceBinary += ".exe"
			}

			// If current exe is different from GOBIN location, copy the new binary
			if currentExe != "" && currentExe != sourceBinary {
				// Check if source exists
				if _, err := os.Stat(sourceBinary); err == nil {
					// Read the new binary
					newBinary, err := os.ReadFile(sourceBinary)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Could not read new binary: %v\n", err)
					} else {
						// Write to current location
						if err := os.WriteFile(currentExe, newBinary, 0755); err != nil {
							fmt.Fprintf(os.Stderr, "Warning: Could not update %s: %v\n", currentExe, err)
							fmt.Fprintf(os.Stderr, "You may need to run the install script again or copy manually.\n")
						} else {
							fmt.Printf("Updated binary at %s\n", currentExe)
						}
					}
				}
			}

			fmt.Println("\nUpgrade complete!")
			fmt.Println("Run 'echonext version' to see the new version.")
		},
	}
}