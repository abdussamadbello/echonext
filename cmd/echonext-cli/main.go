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

			// Get current executable path (resolve symlinks)
			currentExe, err := os.Executable()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not determine current executable path: %v\n", err)
				os.Exit(1)
			}
			// Resolve any symlinks to get the real path
			currentExe, err = filepath.EvalSymlinks(currentExe)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not resolve executable path: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Current binary: %s\n", currentExe)

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

			fmt.Printf("Installing to: %s\n\n", sourceBinary)

			// Run go install to get latest version
			installCmd := exec.Command("go", "install", "github.com/abdussamadbello/echonext/cmd/echonext-cli@latest")
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr

			if err := installCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Upgrade failed: %v\n", err)
				os.Exit(1)
			}

			// Check if source binary was created
			if _, err := os.Stat(sourceBinary); err != nil {
				fmt.Fprintf(os.Stderr, "Error: New binary not found at %s: %v\n", sourceBinary, err)
				os.Exit(1)
			}

			// If current exe is different from GOBIN location, copy the new binary
			if currentExe != sourceBinary {
				fmt.Printf("\nCopying new binary to %s...\n", currentExe)

				// Read the new binary
				newBinary, err := os.ReadFile(sourceBinary)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Could not read new binary: %v\n", err)
					fmt.Fprintf(os.Stderr, "You can manually copy: cp %s %s\n", sourceBinary, currentExe)
					os.Exit(1)
				}

				// On Linux/Unix, we can't overwrite a running binary directly
				// Solution: remove the old binary first (works even if running), then write new one
				if err := os.Remove(currentExe); err != nil {
					fmt.Fprintf(os.Stderr, "Error: Could not remove old binary %s: %v\n", currentExe, err)
					fmt.Fprintf(os.Stderr, "Try with sudo: sudo cp %s %s\n", sourceBinary, currentExe)
					os.Exit(1)
				}

				// Write new binary to the location
				if err := os.WriteFile(currentExe, newBinary, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error: Could not write new binary to %s: %v\n", currentExe, err)
					fmt.Fprintf(os.Stderr, "Try with sudo: sudo cp %s %s\n", sourceBinary, currentExe)
					os.Exit(1)
				}

				fmt.Printf("Updated: %s\n", currentExe)
			}

			fmt.Println("\nUpgrade complete!")
			fmt.Println("Run 'echonext version' to see the new version.")
		},
	}
}