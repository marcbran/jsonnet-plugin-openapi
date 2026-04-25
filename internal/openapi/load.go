package openapi

import "context"

type Loader interface {
	Load(ctx context.Context, ref string) (APISpec, error)
}

type APISpec struct {
	Title   string
	Version string
	Paths   []PathItem
}

type PathItem struct {
	Path       string
	Parameters []Parameter
	Get        *Operation
}

type Operation struct {
	OperationID string
	Parameters  []Parameter
}

type Parameter struct {
	Name     string
	In       string
	Required bool
}
