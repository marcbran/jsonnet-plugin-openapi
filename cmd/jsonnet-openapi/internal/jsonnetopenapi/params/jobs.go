package params

import (
	"context"
	_ "embed"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/inference"
)

const ParamsJobName = "params-inference"

//go:embed schemas/params-inference-output.schema.json
var paramsSchema []byte

type Renderer interface {
	inference.BundleRenderer
	RenderCheck(ctx context.Context, template string, input, output []byte) ([]byte, []string, error)
}

type ParamsJob struct {
	renderer Renderer
}

func NewParamsJob(renderer Renderer) *ParamsJob {
	return &ParamsJob{renderer: renderer}
}

func (j *ParamsJob) Name() string {
	return ParamsJobName
}

func (j *ParamsJob) ValidateOutput(ctx context.Context, task inference.Task, output []byte) ([]byte, []string, error) {
	return j.renderer.RenderCheck(ctx, "params-output-check.jsonnet", task.Input, output)
}

func (j *ParamsJob) Build(ctx context.Context, spec inference.SpecDocument, previous inference.Results) ([]inference.Task, error) {
	bundles, err := j.renderer.RenderBundles(ctx, "params-inference-bundles.jsonnet", spec.JSON, "")
	if err != nil {
		return nil, err
	}
	tasks := make([]inference.Task, 0, len(bundles))
	for _, bundle := range bundles {
		tasks = append(tasks, inference.Task{
			JobName:      j.Name(),
			ID:           bundle.ID,
			Input:        bundle.Input,
			Prompt:       paramsPrompt,
			OutputSchema: paramsSchema,
		})
	}
	return tasks, nil
}
