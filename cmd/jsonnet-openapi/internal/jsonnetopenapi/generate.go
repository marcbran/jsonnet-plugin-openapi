package jsonnetopenapi

import (
	"context"
	"os"

	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
)

type facade struct {
	loader OpenAPILoader
}

func NewFacade(loader OpenAPILoader) openapipkg.Facade {
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
	api, err := g.loader.Load(ctx, in.Spec)
	if err != nil {
		return openapipkg.Output{}, err
	}
	err = os.MkdirAll(in.OutDir, 0755)
	if err != nil {
		return openapipkg.Output{}, err
	}
	nested, err := BuildNestedSpec(api)
	if err != nil {
		return openapipkg.Output{}, err
	}
	service, err := ResolveServiceName(in.Service, api.Title, in.Spec)
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
