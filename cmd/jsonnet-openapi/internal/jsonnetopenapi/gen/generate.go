package gen

import (
	"context"
	"encoding/json"
	"os"

	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/jsonnet-plugin-jsonnet/jsonnet"
	internalopenapi "github.com/marcbran/jsonnet-plugin-openapi/internal/openapi"
)

type Input struct {
	Ref     string
	OutDir  string
	Service string
	PkgRepo string
}

type Output struct {
	OutDir string
	Files  []string
}

func Batch(ctx context.Context, loader internalopenapi.Loader, jobs []Input) ([]Output, error) {
	outs := make([]Output, 0, len(jobs))
	for _, in := range jobs {
		out, err := Generate(ctx, loader, in)
		if err != nil {
			return nil, err
		}
		outs = append(outs, out)
	}
	return outs, nil
}

func Generate(ctx context.Context, loader internalopenapi.Loader, in Input) (Output, error) {
	api, err := loader.Load(ctx, in.Ref)
	if err != nil {
		return Output{}, err
	}
	err = os.MkdirAll(in.OutDir, 0755)
	if err != nil {
		return Output{}, err
	}
	nested, err := internalopenapi.BuildNestedSpec(api)
	if err != nil {
		return Output{}, err
	}
	service, err := internalopenapi.ResolveServiceName(in.Service, api.Title, in.Ref)
	if err != nil {
		return Output{}, err
	}
	err = writeGeneratedLibsonnet(in.OutDir, nested, service, in.PkgRepo)
	if err != nil {
		return Output{}, err
	}
	return Output{
		OutDir: in.OutDir,
		Files: []string{
			"main.libsonnet",
			"pkg.libsonnet",
		},
	}, nil
}

type generationInput struct {
	Spec    *internalopenapi.NestedSpec `json:"spec"`
	Service string                      `json:"service"`
	PkgRepo string                      `json:"pkgRepo"`
}

func writeGeneratedLibsonnet(outDir string, spec *internalopenapi.NestedSpec, service string, pkgRepo string) error {
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
