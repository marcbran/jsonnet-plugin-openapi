package links

import (
	"context"
	_ "embed"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/inference"
)

const LinksFromResourcesJobName = "links-from-resources"
const LinksFromCollectionsJobName = "links-from-collections"
const LinkVarsJobName = "link-vars"

//go:embed schemas/link-inference-output.schema.json
var linkSchema []byte

//go:embed schemas/link-vars-inference-output.schema.json
var linkVarsSchema []byte

type LinksFromResourcesJob struct {
	renderer inference.BundleRenderer
}

func NewLinksFromResourcesJob(renderer inference.BundleRenderer) *LinksFromResourcesJob {
	return &LinksFromResourcesJob{renderer: renderer}
}

func (j *LinksFromResourcesJob) Name() string {
	return LinksFromResourcesJobName
}

func (j *LinksFromResourcesJob) TransformOutput(task inference.Task, output []byte) ([]byte, error) {
	return overrideLinksSourcePath(task.Input, output)
}

func (j *LinksFromResourcesJob) Build(ctx context.Context, spec inference.SpecDocument, previous inference.Results) ([]inference.Task, error) {
	bundles, err := j.renderer.RenderBundles("links-from-resources-bundles.jsonnet", spec.JSON, "")
	if err != nil {
		return nil, err
	}
	tasks := make([]inference.Task, 0, len(bundles))
	for _, bundle := range bundles {
		tasks = append(tasks, inference.Task{
			JobName:      j.Name(),
			ID:           bundle.ID,
			Input:        bundle.Input,
			Prompt:       linksFromResourcesPrompt,
			OutputSchema: linkSchema,
		})
	}
	return tasks, nil
}

type LinksFromCollectionsJob struct {
	renderer inference.BundleRenderer
}

func NewLinksFromCollectionsJob(renderer inference.BundleRenderer) *LinksFromCollectionsJob {
	return &LinksFromCollectionsJob{renderer: renderer}
}

func (j *LinksFromCollectionsJob) Name() string {
	return LinksFromCollectionsJobName
}

func (j *LinksFromCollectionsJob) TransformOutput(task inference.Task, output []byte) ([]byte, error) {
	return overrideLinksSourcePath(task.Input, output)
}

func (j *LinksFromCollectionsJob) Build(ctx context.Context, spec inference.SpecDocument, previous inference.Results) ([]inference.Task, error) {
	bundles, err := j.renderer.RenderBundles("links-from-collections-bundles.jsonnet", spec.JSON, "")
	if err != nil {
		return nil, err
	}
	tasks := make([]inference.Task, 0, len(bundles))
	for _, bundle := range bundles {
		tasks = append(tasks, inference.Task{
			JobName:      j.Name(),
			ID:           bundle.ID,
			Input:        bundle.Input,
			Prompt:       linksFromCollectionsPrompt,
			OutputSchema: linkSchema,
		})
	}
	return tasks, nil
}

type LinkVarsJob struct {
	renderer inference.BundleRenderer
}

func NewLinkVarsJob(renderer inference.BundleRenderer) *LinkVarsJob {
	return &LinkVarsJob{renderer: renderer}
}

func (j *LinkVarsJob) Name() string {
	return LinkVarsJobName
}

func (j *LinkVarsJob) TransformOutput(task inference.Task, output []byte) ([]byte, error) {
	return overrideEchoedFields(task.Input, output, "sourcePath", "targetPath", "at", "keys")
}

func (j *LinkVarsJob) Build(ctx context.Context, spec inference.SpecDocument, previous inference.Results) ([]inference.Task, error) {
	bundles, err := j.renderer.RenderBundles("link-vars-bundles.jsonnet", spec.JSON, mergedLinkResults(previous))
	if err != nil {
		return nil, err
	}
	tasks := make([]inference.Task, 0, len(bundles))
	for _, bundle := range bundles {
		tasks = append(tasks, inference.Task{
			JobName:      j.Name(),
			ID:           bundle.ID,
			Input:        bundle.Input,
			Prompt:       linkVarsPrompt,
			OutputSchema: linkVarsSchema,
		})
	}
	return tasks, nil
}

func mergedLinkResults(previous inference.Results) string {
	return mergeResults(previous[LinksFromResourcesJobName], previous[LinksFromCollectionsJobName])
}
