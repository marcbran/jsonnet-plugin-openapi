package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "openapi-gen",
	Short: "Generate Jsonnet from OpenAPI documents",
}

func init() {
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress progress messages on stderr")
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(batchCmd)
	rootCmd.Version = version
}

func Execute() {
	err := rootCmd.Execute()
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
