package kinopenapi

import (
	"context"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Load(ctx context.Context, ref string) (jsonnetopenapi.APISpec, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.Context = ctx
	var doc *openapi3.T
	var err error
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		u, perr := url.Parse(ref)
		if perr != nil {
			return jsonnetopenapi.APISpec{}, perr
		}
		doc, err = loader.LoadFromURI(u)
	} else {
		abs, perr := filepath.Abs(ref)
		if perr != nil {
			return jsonnetopenapi.APISpec{}, perr
		}
		doc, err = loader.LoadFromFile(abs)
	}
	if err != nil {
		return jsonnetopenapi.APISpec{}, err
	}
	err = doc.Validate(ctx, openapi3.DisableExamplesValidation())
	if err != nil {
		return jsonnetopenapi.APISpec{}, err
	}
	api, err := mapDocument(doc)
	if err != nil {
		return jsonnetopenapi.APISpec{}, err
	}
	return api, nil
}

func mapDocument(doc *openapi3.T) (jsonnetopenapi.APISpec, error) {
	var api jsonnetopenapi.APISpec
	if doc.Info != nil {
		api.Title = doc.Info.Title
		api.Version = doc.Info.Version
	}
	if doc.Paths == nil {
		return api, nil
	}
	for path, item := range doc.Paths.Map() {
		if item == nil || item.Get == nil {
			continue
		}
		pathParams := make([]jsonnetopenapi.Parameter, 0, len(item.Parameters))
		for _, p := range item.Parameters {
			if p == nil {
				continue
			}
			if p.Value == nil {
				continue
			}
			pathParams = append(pathParams, jsonnetopenapi.Parameter{
				Name:     p.Value.Name,
				In:       string(p.Value.In),
				Required: p.Value.Required,
			})
		}
		getParams := make([]jsonnetopenapi.Parameter, 0, len(item.Get.Parameters))
		for _, p := range item.Get.Parameters {
			if p == nil {
				continue
			}
			if p.Value == nil {
				continue
			}
			getParams = append(getParams, jsonnetopenapi.Parameter{
				Name:     p.Value.Name,
				In:       string(p.Value.In),
				Required: p.Value.Required,
			})
		}
		api.Paths = append(api.Paths, jsonnetopenapi.PathItem{
			Path:       path,
			Parameters: pathParams,
			Get: &jsonnetopenapi.Operation{
				OperationID: item.Get.OperationID,
				Parameters:  getParams,
			},
		})
	}
	return api, nil
}
