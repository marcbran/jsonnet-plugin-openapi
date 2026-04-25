package jsonnetopenapi

import (
	"encoding/json"

	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/jsonnet-plugin-jsonnet/jsonnet"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi/lib/imports"
)

func writeGeneratedLibsonnet(outDir string, spec *NestedSpec, service string, pkgRepo string) error {
	apiJSON, err := json.Marshal(generationInput{
		Spec:    spec,
		Service: service,
		PkgRepo: pkgRepo,
	})
	if err != nil {
		return err
	}
	err = jpoet.Eval(
		jpoet.FileImport([]string{}),
		jpoet.FSImport(lib),
		jpoet.FSImport(imports.Fs),
		jpoet.WithPlugin(jsonnet.Plugin()),
		jpoet.TLACode("api", string(apiJSON)),
		jpoet.FileInput("./lib/gen.libsonnet"),
		jpoet.Serialize(false),
		jpoet.DirectoryOutput(outDir),
	)
	if err != nil {
		return err
	}
	return nil
}

type generationInput struct {
	Spec    *NestedSpec `json:"spec"`
	Service string      `json:"service"`
	PkgRepo string      `json:"pkgRepo"`
}
