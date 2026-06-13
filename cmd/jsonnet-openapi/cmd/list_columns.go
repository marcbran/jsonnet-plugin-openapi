package cmd

import (
	"fmt"
	"os"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi/listcolumns"
	"github.com/spf13/cobra"
)

var listColumnsCmd = &cobra.Command{
	Use:   "list-columns",
	Short: "Work with OpenAPI list table columns",
}

var listColumnsInferCmd = &cobra.Command{
	Use:   "infer SPEC",
	Short: "Infer table columns for OpenAPI list endpoints",
	Args:  cobra.ExactArgs(1),
	RunE:  runListColumnsInfer,
}

func init() {
	listColumnsInferCmd.Flags().StringP("out", "o", "", "output columns JSON file (default: next to SPEC as <name>.columns.json)")
	listColumnsInferCmd.Flags().String("workdir", "", "working directory for inference bundles and cached results (default: next to SPEC as <name>/)")
	listColumnsInferCmd.Flags().String("model", "", "Codex model to use for inference (default: gpt-5.5)")
	listColumnsInferCmd.Flags().Int("limit", 0, "maximum number of new Codex calls per inference step; 0 means unlimited")
	listColumnsInferCmd.Flags().Bool("force", false, "re-run Codex calls even when result files already exist")
	listColumnsCmd.AddCommand(listColumnsInferCmd)
}

func runListColumnsInfer(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	out, err := cmd.Flags().GetString("out")
	if err != nil {
		return err
	}
	workDir, err := cmd.Flags().GetString("workdir")
	if err != nil {
		return err
	}
	model, err := cmd.Flags().GetString("model")
	if err != nil {
		return err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return err
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}

	if !quiet {
		_, err = fmt.Fprintln(os.Stderr, "inferring columns")
		if err != nil {
			return err
		}
	}

	result, err := listcolumns.Exec(cmd.Context(), listcolumns.Input{
		Spec:    args[0],
		Out:     out,
		WorkDir: workDir,
		Model:   model,
		Limit:   limit,
		Force:   force,
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(os.Stdout, result.Out)
	return err
}
