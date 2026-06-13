package jsonnetopenapi

import "context"

type Input struct {
	Ref     string `json:"ref"`
	OutDir  string `json:"outDir"`
	Service string `json:"service,omitempty"`
	PkgRepo string `json:"pkgRepo,omitempty"`
}

type Output struct {
	OutDir string   `json:"outDir"`
	Files  []string `json:"files"`
}

type ListDetailLinksInput struct {
	Spec    string `json:"spec"`
	Out     string `json:"out,omitempty"`
	WorkDir string `json:"workDir,omitempty"`
	Model   string `json:"model,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Force   bool   `json:"force,omitempty"`
}

type ListDetailLinksOutput struct {
	Out     string   `json:"out"`
	WorkDir string   `json:"workDir"`
	Files   []string `json:"files"`
}

type Facade interface {
	Generate(ctx context.Context, in Input) (Output, error)
	Batch(ctx context.Context, jobs []Input) ([]Output, error)
	InferListDetailLinks(ctx context.Context, in ListDetailLinksInput) (ListDetailLinksOutput, error)
}
