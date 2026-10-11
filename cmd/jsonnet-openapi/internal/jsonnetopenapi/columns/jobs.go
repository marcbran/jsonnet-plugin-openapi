package columns

import (
	"context"
	_ "embed"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/inference"
)

const ColumnsJobName = "columns-inference"

//go:embed schemas/columns-inference-output.schema.json
var columnsSchema []byte

type Renderer interface {
	inference.BundleRenderer
	RenderCheck(template string, input, output []byte) ([]byte, []string, error)
}

type ColumnsJob struct {
	renderer Renderer
}

func NewColumnsJob(renderer Renderer) *ColumnsJob {
	return &ColumnsJob{renderer: renderer}
}

func (j *ColumnsJob) Name() string {
	return ColumnsJobName
}

func (j *ColumnsJob) ValidateOutput(task inference.Task, output []byte) ([]byte, []string, error) {
	return j.renderer.RenderCheck("columns-output-check.jsonnet", task.Input, output)
}

func (j *ColumnsJob) Build(ctx context.Context, spec inference.SpecDocument, previous inference.Results) ([]inference.Task, error) {
	bundles, err := j.renderer.RenderBundles("columns-inference-bundles.jsonnet", spec.JSON, "")
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
