package jsonnetopenapi

import (
	"context"
	"encoding/json"
	"os"

	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/jsonnet-plugin-jsonnet/jsonnet"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi/lib/imports"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi/listdetaillinks"
	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	internalopenapi "github.com/marcbran/jsonnet-plugin-openapi/internal/openapi"
)

type facade struct {
	loader internalopenapi.Loader
}

func NewFacade(loader internalopenapi.Loader) openapipkg.Facade {
	return &facade{loader: loader}
}

func (g *facade) Batch(ctx context.Context, jobs []openapipkg.Input) ([]openapipkg.Output, error) {
	outs := make([]openapipkg.Output, 0, len(jobs))
	for _, in := range jobs {
		out, err := g.Generate(ctx, in)
		if err != nil {
			return nil, err
		}
		outs = append(outs, out)
	}
	return outs, nil
}

func (g *facade) Generate(ctx context.Context, in openapipkg.Input) (openapipkg.Output, error) {
	api, err := g.loader.Load(ctx, in.Ref)
	if err != nil {
		return openapipkg.Output{}, err
	}
	err = os.MkdirAll(in.OutDir, 0755)
	if err != nil {
		return openapipkg.Output{}, err
	}
	nested, err := internalopenapi.BuildNestedSpec(api)
	if err != nil {
		return openapipkg.Output{}, err
	}
	service, err := internalopenapi.ResolveServiceName(in.Service, api.Title, in.Ref)
	if err != nil {
		return openapipkg.Output{}, err
	}
	err = writeGeneratedLibsonnet(in.OutDir, nested, service, in.PkgRepo)
	if err != nil {
		return openapipkg.Output{}, err
	}
	return openapipkg.Output{
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

func (g *facade) InferListDetailLinks(ctx context.Context, in openapipkg.ListDetailLinksInput) (openapipkg.ListDetailLinksOutput, error) {
	out, err := listdetaillinks.Exec(ctx, listdetaillinks.Input{
		Spec:    in.Spec,
		Out:     in.Out,
		WorkDir: in.WorkDir,
		Model:   in.Model,
		Limit:   in.Limit,
		Force:   in.Force,
	})
	if err != nil {
		return openapipkg.ListDetailLinksOutput{}, err
	}
	return openapipkg.ListDetailLinksOutput{
		Out:     out.Out,
		WorkDir: out.WorkDir,
		Files:   out.Files,
	}, nil
}
