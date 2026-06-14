package listcolumns

import (
	"context"
	_ "embed"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/infra/inference"
)

const ColumnsJobName = "list-columns-inference"

//go:embed schemas/list-columns-inference-output.schema.json
var columnsSchema []byte

type ColumnsJob struct {
	renderer inference.BundleRenderer
}

func NewColumnsJob(renderer inference.BundleRenderer) *ColumnsJob {
	return &ColumnsJob{renderer: renderer}
}

func (j *ColumnsJob) Name() string {
	return ColumnsJobName
}

func (j *ColumnsJob) Build(ctx context.Context, spec inference.SpecDocument, previous inference.Results) ([]inference.Task, error) {
	bundles, err := j.renderer.RenderBundles("list-columns-inference-bundles.jsonnet", spec.JSON, "")
	if err != nil {
		return nil, err
	}
	tasks := make([]inference.Task, 0, len(bundles))
	for _, bundle := range bundles {
		tasks = append(tasks, inference.Task{
			JobName:      j.Name(),
			ID:           bundle.ID,
			Input:        bundle.Input,
			Prompt:       columnsPrompt,
			OutputSchema: columnsSchema,
		})
	}
	return tasks, nil
}
