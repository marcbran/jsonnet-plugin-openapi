package jsonnetopenapi

import (
	"context"
	"os"
	"path/filepath"

	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
)

const resolvedSpecFile = "openapi.resolved.json"

type facade struct {
	loader OpenAPILoader
}

func NewFacade(loader OpenAPILoader) openapipkg.Facade {
	return &facade{loader: loader}
}

func (g *facade) Generate(ctx context.Context, in openapipkg.Input) (openapipkg.Output, error) {
	ls, err := g.loader.Load(ctx, in.Spec)
	if err != nil {
		return openapipkg.Output{}, err
	}
	err = os.MkdirAll(in.OutDir, 0755)
	if err != nil {
		return openapipkg.Output{}, err
	}
	resPath := filepath.Join(in.OutDir, resolvedSpecFile)
	err = os.WriteFile(resPath, ls.ResolvedJSON, 0666)
	if err != nil {
		return openapipkg.Output{}, err
	}
	payload, err := BuildPayload(ls.API, in.Service, in.Spec)
	if err != nil {
		return openapipkg.Output{}, err
	}
	payload.PkgRepo = in.PkgRepo
	err = writeGeneratedLibsonnet(in.OutDir, payload)
	if err != nil {
		return openapipkg.Output{}, err
	}
	return openapipkg.Output{
		OutDir: in.OutDir,
		Files: []string{
			resolvedSpecFile,
			"main.libsonnet",
			"pkg.libsonnet",
		},
	}, nil
}
