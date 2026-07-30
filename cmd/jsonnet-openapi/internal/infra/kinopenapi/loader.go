package kinopenapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/inference"
	"sigs.k8s.io/yaml"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) LoadSpec(ctx context.Context, ref string) (inference.SpecDocument, error) {
	data, err := readRef(ref)
	if err != nil {
		return inference.SpecDocument{}, err
	}
	location, err := refLocation(ref)
	if err != nil {
		return inference.SpecDocument{}, err
	}
	if isSwagger2(data) {
		return loadSwagger2Spec(ctx, data, location)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.Context = ctx
	doc, err := loader.LoadFromDataWithPath(data, location)
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

// refLocation builds the base URL used to resolve relative external $refs,
// mirroring how openapi3.Loader.LoadFromFile/LoadFromURI derive it.
func refLocation(ref string) (*url.URL, error) {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return url.Parse(ref)
	}
	abs, err := filepath.Abs(ref)
	if err != nil {
		return nil, err
	}
	return &url.URL{Path: filepath.ToSlash(abs)}, nil
}

func isSwagger2(data []byte) bool {
	var probe struct {
		Swagger string `json:"swagger"`
	}
	if err := yaml.Unmarshal(data, &probe); err != nil {
		return false
	}
	return strings.HasPrefix(probe.Swagger, "2.")
}

func readRef(ref string) ([]byte, error) {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		resp, err := http.Get(ref)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()
		return io.ReadAll(resp.Body)
	}
	abs, err := filepath.Abs(ref)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(abs)
}

func loadSwagger2Spec(ctx context.Context, data []byte, location *url.URL) (inference.SpecDocument, error) {
	var doc2 openapi2.T
	if err := yaml.Unmarshal(data, &doc2); err != nil {
		return inference.SpecDocument{}, fmt.Errorf("parsing swagger 2.0 document: %w", err)
	}
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.Context = ctx
	doc3, err := openapi2conv.ToV3WithLoader(&doc2, loader, location)
	if err != nil {
		return inference.SpecDocument{}, fmt.Errorf("converting swagger 2.0 to openapi 3: %w", err)
	}
	doc3.InternalizeRefs(ctx, nil)
	raw, err := json.Marshal(doc3)
	if err != nil {
		return inference.SpecDocument{}, err
	}
	return inference.SpecDocument{JSON: string(raw)}, nil
}
