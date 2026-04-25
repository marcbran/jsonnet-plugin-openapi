package jsonnetopenapi

import "context"

type OpenAPILoader interface {
	Load(ctx context.Context, ref string) (APISpec, error)
}

type APISpec struct {
	Title         string
	Version       string
	GETOperations []GETOperation
}

type GETOperation struct {
	Path        string
	OperationID string
	Parameters  []Parameter
}

type Parameter struct {
	Name     string
	In       string
	Required bool
}
