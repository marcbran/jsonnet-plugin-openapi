package jsonnetopenapi

import "context"

type Input struct {
	Spec    string `json:"spec"`
	OutDir  string `json:"outDir"`
	Service string `json:"service,omitempty"`
	PkgRepo string `json:"pkgRepo,omitempty"`
}

type Output struct {
	OutDir string   `json:"outDir"`
	Files  []string `json:"files"`
}

type Facade interface {
	Generate(ctx context.Context, in Input) (Output, error)
	Batch(ctx context.Context, jobs []Input) ([]Output, error)
}
