package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev" // Will be set during build

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
			fmt.Printf("EchoNext CLI %s\n", version)
			fmt.Println("Build type-safe APIs with automatic OpenAPI generation")
			fmt.Println("https://github.com/abdussamadbello/echonext")
		},
	}
}