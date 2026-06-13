package cmd

import (
	"fmt"
	"os"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi/inferlinks"
	"github.com/spf13/cobra"
)

var inferLinksCmd = &cobra.Command{
	Use:   "infer-links SPEC",
	Short: "Infer list-to-detail links for an OpenAPI JSON document",
	Args:  cobra.ExactArgs(1),
	RunE:  runInferLinks,
}

func init() {
	inferLinksCmd.Flags().StringP("out", "o", "", "output links JSON file (default: next to SPEC as <name>.links.json)")
	inferLinksCmd.Flags().String("workdir", "", "working directory for inference bundles and cached results (default: next to SPEC as <name>/)")
	inferLinksCmd.Flags().String("model", "", "Codex model to use for inference (default: gpt-5.5)")
	inferLinksCmd.Flags().Int("limit", 0, "maximum number of new Codex calls per inference step; 0 means unlimited")
	inferLinksCmd.Flags().Bool("force", false, "re-run Codex calls even when result files already exist")
}

func runInferLinks(cmd *cobra.Command, args []string) error {
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
		_, err = fmt.Fprintln(os.Stderr, "inferring links")
		if err != nil {
			return err
		}
	}

	result, err := inferlinks.Run(cmd.Context(), inferlinks.Input{
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
