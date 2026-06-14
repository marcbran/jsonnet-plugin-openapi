package listdetaillinks

import (
	"context"
	_ "embed"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/infra/inference"
)

const ListDetailJobName = "list-detail-inference"
const VarsJobName = "list-detail-vars-inference"

//go:embed schemas/list-detail-inference-output.schema.json
var listDetailSchema []byte

//go:embed schemas/list-detail-vars-inference-output.schema.json
var varsSchema []byte

type ListDetailJob struct {
	renderer inference.BundleRenderer
}

func NewListDetailJob(renderer inference.BundleRenderer) *ListDetailJob {
	return &ListDetailJob{renderer: renderer}
}

func (j *ListDetailJob) Name() string {
	return ListDetailJobName
}

func (j *ListDetailJob) Build(ctx context.Context, spec inference.SpecDocument, previous inference.Results) ([]inference.Task, error) {
	bundles, err := j.renderer.RenderBundles("list-detail-inference-bundles.jsonnet", spec.JSON, "")
	if err != nil {
		return nil, err
	}
	tasks := make([]inference.Task, 0, len(bundles))
	for _, bundle := range bundles {
		tasks = append(tasks, inference.Task{
			JobName:      j.Name(),
			ID:           bundle.ID,
			Input:        bundle.Input,
			Prompt:       listDetailPrompt,
			OutputSchema: listDetailSchema,
		})
	}
	return tasks, nil
}

type VarsJob struct {
	renderer inference.BundleRenderer
}

func NewVarsJob(renderer inference.BundleRenderer) *VarsJob {
	return &VarsJob{renderer: renderer}
}

func (j *VarsJob) Name() string {
	return VarsJobName
}

func (j *VarsJob) Build(ctx context.Context, spec inference.SpecDocument, previous inference.Results) ([]inference.Task, error) {
	bundles, err := j.renderer.RenderBundles("list-detail-vars-inference-bundles.jsonnet", spec.JSON, previous[ListDetailJobName])
	if err != nil {
		return nil, err
	}
	tasks := make([]inference.Task, 0, len(bundles))
	for _, bundle := range bundles {
		tasks = append(tasks, inference.Task{
			JobName:      j.Name(),
			ID:           bundle.ID,
			Input:        bundle.Input,
			Prompt:       varsPrompt,
			OutputSchema: varsSchema,
		})
	}
	return tasks, nil
}
