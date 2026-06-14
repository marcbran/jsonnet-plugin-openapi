package kinopenapi

import (
	"context"
	"encoding/json"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/inference"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) LoadSpec(ctx context.Context, ref string) (inference.SpecDocument, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.Context = ctx
	var doc *openapi3.T
	var err error
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		u, perr := url.Parse(ref)
		if perr != nil {
			return inference.SpecDocument{}, perr
		}
		doc, err = loader.LoadFromURI(u)
	} else {
		abs, perr := filepath.Abs(ref)
		if perr != nil {
			return inference.SpecDocument{}, perr
		}
		doc, err = loader.LoadFromFile(abs)
	}
	if err != nil {
		return inference.SpecDocument{}, err
	}
	doc.InternalizeRefs(ctx, nil)
	raw, err := json.Marshal(doc)
	if err != nil {
		return inference.SpecDocument{}, err
	}
	return inference.SpecDocument{JSON: string(raw)}, nil
}
