package jsonnetopenapi

import "context"

type OpenAPILoader interface {
	Load(ctx context.Context, ref string) (LoadedSpec, error)
}

type LoadedSpec struct {
	API          APISpec
	ResolvedJSON []byte
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
