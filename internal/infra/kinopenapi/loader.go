package kinopenapi

import (
	"context"
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
	"sigs.k8s.io/yaml"

	internalopenapi "github.com/marcbran/jsonnet-plugin-openapi/internal/openapi"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Parse(ctx context.Context, spec string) (internalopenapi.APISpec, error) {
	data := []byte(spec)
	if isSwagger2(data) {
		doc, err := loadSwagger2Data(ctx, data, nil)
		if err != nil {
			return internalopenapi.APISpec{}, err
		}
		return finishLoad(ctx, doc)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.Context = ctx
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return internalopenapi.APISpec{}, err
	}
	return finishLoad(ctx, doc)
}

func (l *Loader) Load(ctx context.Context, ref string) (internalopenapi.APISpec, error) {
	data, err := readRef(ref)
	if err != nil {
		return internalopenapi.APISpec{}, err
	}
	location, err := refLocation(ref)
	if err != nil {
		return internalopenapi.APISpec{}, err
	}
	if isSwagger2(data) {
		doc, err := loadSwagger2Data(ctx, data, location)
		if err != nil {
			return internalopenapi.APISpec{}, err
		}
		return finishLoad(ctx, doc)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.Context = ctx
	doc, err := loader.LoadFromDataWithPath(data, location)
	if err != nil {
		return internalopenapi.APISpec{}, err
	}
	return finishLoad(ctx, doc)
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

func finishLoad(ctx context.Context, doc *openapi3.T) (internalopenapi.APISpec, error) {
	pruneNonGETPaths(doc)
	err := doc.Validate(ctx, openapi3.DisableExamplesValidation())
	if err != nil {
		return internalopenapi.APISpec{}, err
	}
	return mapDocument(doc)
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

func loadSwagger2Data(ctx context.Context, data []byte, location *url.URL) (*openapi3.T, error) {
	var doc2 openapi2.T
	if err := yaml.Unmarshal(data, &doc2); err != nil {
		return nil, fmt.Errorf("parsing swagger 2.0 document: %w", err)
	}
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.Context = ctx
	doc3, err := openapi2conv.ToV3WithLoader(&doc2, loader, location)
	if err != nil {
		return nil, fmt.Errorf("converting swagger 2.0 to openapi 3: %w", err)
	}
	return doc3, nil
}

// pruneNonGETPaths drops path items with no GET operation before validation.
// mapDocument only ever reads item.Get, so these are dead weight regardless -
// pruning them first also avoids spurious path-uniqueness conflicts between
// a GET path and a non-GET path that share the same shape but differently
// named parameters.
func pruneNonGETPaths(doc *openapi3.T) {
	if doc.Paths == nil {
		return
	}
	for path, item := range doc.Paths.Map() {
		if item == nil || item.Get == nil {
			doc.Paths.Delete(path)
		}
	}
}

func mapDocument(doc *openapi3.T) (internalopenapi.APISpec, error) {
	var api internalopenapi.APISpec
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
		pathParams := make([]internalopenapi.Parameter, 0, len(item.Parameters))
		for _, p := range item.Parameters {
			if p == nil || p.Value == nil {
				continue
			}
			pathParams = append(pathParams, internalopenapi.Parameter{
				Name:     p.Value.Name,
				In:       string(p.Value.In),
				Required: p.Value.Required,
			})
		}
		getParams := make([]internalopenapi.Parameter, 0, len(item.Get.Parameters))
		for _, p := range item.Get.Parameters {
			if p == nil || p.Value == nil {
				continue
			}
			getParams = append(getParams, internalopenapi.Parameter{
				Name:     p.Value.Name,
				In:       string(p.Value.In),
				Required: p.Value.Required,
			})
		}
		api.Paths = append(api.Paths, internalopenapi.PathItem{
			Path:       path,
			Parameters: pathParams,
			Get: &internalopenapi.Operation{
				OperationID: item.Get.OperationID,
				Parameters:  getParams,
			},
		})
	}
	return api, nil
}
