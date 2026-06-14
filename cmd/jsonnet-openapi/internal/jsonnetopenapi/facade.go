package jsonnetopenapi

import (
	"context"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi/gen"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi/listcolumns"
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
	ins := make([]gen.Input, 0, len(jobs))
	for _, in := range jobs {
		ins = append(ins, gen.Input{
			Ref:     in.Ref,
			OutDir:  in.OutDir,
			Service: in.Service,
			PkgRepo: in.PkgRepo,
		})
	}
	outs, err := gen.Batch(ctx, g.loader, ins)
	if err != nil {
		return nil, err
	}
	result := make([]openapipkg.Output, 0, len(outs))
	for _, out := range outs {
		result = append(result, openapipkg.Output{
			OutDir: out.OutDir,
			Files:  out.Files,
		})
	}
	return result, nil
}

func (g *facade) Generate(ctx context.Context, in openapipkg.Input) (openapipkg.Output, error) {
	out, err := gen.Generate(ctx, g.loader, gen.Input{
		Ref:     in.Ref,
		OutDir:  in.OutDir,
		Service: in.Service,
		PkgRepo: in.PkgRepo,
	})
	if err != nil {
		return openapipkg.Output{}, err
	}
	return openapipkg.Output{
		OutDir: out.OutDir,
		Files:  out.Files,
	}, nil
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

func (g *facade) InferListColumns(ctx context.Context, in openapipkg.ListColumnsInput) (openapipkg.ListColumnsOutput, error) {
	out, err := listcolumns.Exec(ctx, listcolumns.Input{
		Spec:    in.Spec,
		Out:     in.Out,
		WorkDir: in.WorkDir,
		Model:   in.Model,
		Limit:   in.Limit,
		Force:   in.Force,
	})
	if err != nil {
		return openapipkg.ListColumnsOutput{}, err
	}
	return openapipkg.ListColumnsOutput{
		Out:     out.Out,
		WorkDir: out.WorkDir,
		Files:   out.Files,
	}, nil
}
