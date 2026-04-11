package kinopenapi

import (
	"context"
	"encoding/json"
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

func (l *Loader) Load(ctx context.Context, ref string) (jsonnetopenapi.LoadedSpec, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.Context = ctx
	var doc *openapi3.T
	var err error
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		u, perr := url.Parse(ref)
		if perr != nil {
			return jsonnetopenapi.LoadedSpec{}, perr
		}
		doc, err = loader.LoadFromURI(u)
	} else {
		abs, perr := filepath.Abs(ref)
		if perr != nil {
			return jsonnetopenapi.LoadedSpec{}, perr
		}
		doc, err = loader.LoadFromFile(abs)
	}
	if err != nil {
		return jsonnetopenapi.LoadedSpec{}, err
	}
	err = doc.Validate(ctx, openapi3.DisableExamplesValidation())
	if err != nil {
		return jsonnetopenapi.LoadedSpec{}, err
	}
	resolved, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return jsonnetopenapi.LoadedSpec{}, err
	}
	api, err := mapDocument(doc)
	if err != nil {
		return jsonnetopenapi.LoadedSpec{}, err
	}
	return jsonnetopenapi.LoadedSpec{
		API:          api,
		ResolvedJSON: resolved,
	}, nil
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
		params := mergeParameters(item, item.Get)
		domainParams := make([]jsonnetopenapi.Parameter, 0, len(params))
		for _, p := range params {
			if p == nil {
				continue
			}
			domainParams = append(domainParams, jsonnetopenapi.Parameter{
				Name:     p.Name,
				In:       string(p.In),
				Required: p.Required,
			})
		}
		api.GETOperations = append(api.GETOperations, jsonnetopenapi.GETOperation{
			Path:        path,
			OperationID: item.Get.OperationID,
			Parameters:  domainParams,
		})
	}
	return api, nil
}

func mergeParameters(pathItem *openapi3.PathItem, op *openapi3.Operation) []*openapi3.Parameter {
	byKey := map[string]*openapi3.Parameter{}
	add := func(refs openapi3.Parameters) {
		for _, ref := range refs {
			if ref == nil || ref.Value == nil {
				continue
			}
			p := ref.Value
			k := p.In + ":" + p.Name
			byKey[k] = p
		}
	}
	if pathItem != nil {
		add(pathItem.Parameters)
	}
	if op != nil {
		add(op.Parameters)
	}
	out := make([]*openapi3.Parameter, 0, len(byKey))
	for _, p := range byKey {
		out = append(out, p)
	}
	return out
}
