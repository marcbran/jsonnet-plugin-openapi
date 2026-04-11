package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/infra/kinopenapi"
	jnogen "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen SPEC",
	Short: "Generate Jsonnet client from an OpenAPI 3 document",
	Args:  cobra.ExactArgs(1),
	RunE:  runGen,
}

func init() {
	genCmd.Flags().StringP("format", "f", "text", "Output format: text, json")
	genCmd.Flags().StringP("out", "o", ".", "output directory")
	genCmd.Flags().String("service", "", "service slug for std.native('invoke:…') and pkg.libsonnet (from --service, else info.title, else spec filename)")
	genCmd.Flags().String("pkg-repo", "", "git remote URL for generated pkg.libsonnet repo field")
}

func runGen(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}
	outDir, err := cmd.Flags().GetString("out")
	if err != nil {
		return err
	}
	service, err := cmd.Flags().GetString("service")
	if err != nil {
		return err
	}
	pkgRepo, err := cmd.Flags().GetString("pkg-repo")
	if err != nil {
		return err
	}
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}

	if !quiet && format == "text" {
		_, err = fmt.Fprintln(os.Stderr, "generating")
		if err != nil {
			return err
		}
	}

	g := jnogen.NewFacade(kinopenapi.NewLoader())
	out, err := g.Generate(cmd.Context(), jsonnetopenapi.Input{
		Spec:    args[0],
		OutDir:  outDir,
		Service: service,
		PkgRepo: pkgRepo,
	})
	if err != nil {
		return err
	}

	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(out)
	default:
		for _, name := range out.Files {
			_, err = fmt.Fprintln(os.Stdout, filepath.Join(out.OutDir, name))
			if err != nil {
				return err
			}
		}
		return nil
	}
}
