package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"

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

			// Run go install to get latest version
			installCmd := exec.Command("go", "install", "github.com/abdussamadbello/echonext/cmd/echonext-cli@latest")
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr

			if err := installCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Upgrade failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("\nUpgrade complete!")
			fmt.Println("Run 'echonext version' to see the new version.")
		},
	}
}