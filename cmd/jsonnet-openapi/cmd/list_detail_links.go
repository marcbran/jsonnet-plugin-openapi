package cmd

import (
	"fmt"
	"os"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi/listdetaillinks"
	"github.com/spf13/cobra"
)

var listDetailLinksCmd = &cobra.Command{
	Use:   "list-detail-links",
	Short: "Work with list-to-detail OpenAPI links",
}

var listDetailLinksInferCmd = &cobra.Command{
	Use:   "infer SPEC",
	Short: "Infer list-to-detail links for an OpenAPI document",
	Args:  cobra.ExactArgs(1),
	RunE:  runListDetailLinksInfer,
}

func init() {
	listDetailLinksInferCmd.Flags().StringP("out", "o", "", "output links JSON file (default: next to SPEC as <name>.links.json)")
	listDetailLinksInferCmd.Flags().String("workdir", "", "working directory for inference bundles and cached results (default: next to SPEC as <name>/)")
	listDetailLinksInferCmd.Flags().String("model", "", "Codex model to use for inference (default: gpt-5.5)")
	listDetailLinksInferCmd.Flags().Int("limit", 0, "maximum number of new Codex calls per inference step; 0 means unlimited")
	listDetailLinksInferCmd.Flags().Bool("force", false, "re-run Codex calls even when result files already exist")
	listDetailLinksCmd.AddCommand(listDetailLinksInferCmd)
}

func runListDetailLinksInfer(cmd *cobra.Command, args []string) error {
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

	var progress listdetaillinks.Progress
	if !quiet {
		progress = func(jobName, taskID string, cached bool) {
			if cached {
				fmt.Fprintf(os.Stderr, "  [cached]  %s/%s\n", jobName, taskID)
			} else {
				fmt.Fprintf(os.Stderr, "  [running] %s/%s\n", jobName, taskID)
			}
		}
	}

	result, err := listdetaillinks.Exec(cmd.Context(), listdetaillinks.Input{
		Spec:     args[0],
		Out:      out,
		WorkDir:  workDir,
		Model:    model,
		Limit:    limit,
		Force:    force,
		Progress: progress,
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(os.Stdout, result.Out)
	return err
}
