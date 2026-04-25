package openapi

import "context"

type Loader interface {
	Load(ctx context.Context, ref string) (APISpec, error)
}

type Parser interface {
	Parse(ctx context.Context, spec string) (APISpec, error)
}

type APISpec struct {
	Title   string     `json:"title"`
	Version string     `json:"version"`
	Paths   []PathItem `json:"paths"`
}

type PathItem struct {
	Path       string      `json:"path"`
	Parameters []Parameter `json:"parameters,omitempty"`
	Get        *Operation  `json:"get,omitempty"`
}

type Operation struct {
	OperationID string      `json:"operationId"`
	Parameters  []Parameter `json:"parameters,omitempty"`
}

type Parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
}
