package jsonnetopenapi

import "context"

type Input struct {
	Spec    string
	OutDir  string
	Service string
	PkgRepo string
}

type Output struct {
	OutDir string   `json:"outDir"`
	Files  []string `json:"files"`
}

type Facade interface {
	Generate(ctx context.Context, in Input) (Output, error)
}
