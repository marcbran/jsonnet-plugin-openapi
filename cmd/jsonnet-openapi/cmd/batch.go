package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	jnogen "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	"github.com/marcbran/jsonnet-plugin-openapi/internal/infra/kinopenapi"
	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch CONFIG",
	Short: "Run multiple codegen jobs from a JSON config file",
	Args:  cobra.ExactArgs(1),
	RunE:  runBatch,
}

func init() {
	batchCmd.Flags().StringP("format", "f", "text", "Output format: text, json")
}

func runBatch(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	configPath := args[0]
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var jobs []jsonnetopenapi.Input
	err = json.Unmarshal(raw, &jobs)
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	for i := range jobs {
		if jobs[i].Ref != "" && !filepath.IsAbs(jobs[i].Ref) {
			jobs[i].Ref = filepath.Join(configDir, jobs[i].Ref)
		}
		outDir := jobs[i].OutDir
		if outDir == "" {
			outDir = "."
		}
		if !filepath.IsAbs(outDir) {
			jobs[i].OutDir = filepath.Join(configDir, outDir)
		}
	}

	if !quiet && format == "text" {
		_, err = fmt.Fprintln(os.Stderr, "generating")
		if err != nil {
			return err
		}
	}

	g := jnogen.NewFacade(kinopenapi.NewLoader())
	outs, err := g.Batch(cmd.Context(), jobs)
	if err != nil {
		return err
	}

	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(outs)
	default:
		for _, out := range outs {
			for _, name := range out.Files {
				_, err = fmt.Fprintln(os.Stdout, filepath.Join(out.OutDir, name))
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}
