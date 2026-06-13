package jsonnet

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/infra/inference"
)

type Renderer struct {
	fs fs.FS
}

func NewRenderer(fs fs.FS) *Renderer {
	return &Renderer{fs: fs}
}

func (r *Renderer) RenderBundles(template string, specJSON string, previousJSON string) ([]inference.Bundle, error) {
	dir, err := os.MkdirTemp("", "jsonnet-openapi-inference-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	opts := []jpoet.Option{
		jpoet.FSImport(r.fs),
		jpoet.SnippetInput(template, fmt.Sprintf("local spec = %s; (import %q)(spec)", specJSON, template)),
		jpoet.Serialize(false),
		jpoet.DirectoryOutput(dir),
	}
	if previousJSON != "" {
		opts[1] = jpoet.SnippetInput(template, fmt.Sprintf(
			"local spec = %s; local inferred = %s; (import %q)(spec, inferred)",
			specJSON,
			previousJSON,
			template,
		))
	}
	err = jpoet.Eval(opts...)
	if err != nil {
		return nil, err
	}

	inputs, err := filepath.Glob(filepath.Join(dir, "*", "input.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(inputs)
	bundles := make([]inference.Bundle, 0, len(inputs))
	for _, input := range inputs {
		raw, err := os.ReadFile(input)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, inference.Bundle{
			ID:    filepath.Base(filepath.Dir(input)),
			Input: raw,
		})
	}
	return bundles, nil
}

func (r *Renderer) RenderLinks(specJSON string, inferredJSON string, varsJSON string) ([]byte, error) {
	var out bytes.Buffer
	err := jpoet.Eval(
		jpoet.FSImport(r.fs),
		jpoet.SnippetInput("list-detail-links", fmt.Sprintf(
			"local spec = %s; local inferred = %s; local varsInferred = %s; (import 'list-detail-links.jsonnet')(spec, inferred, varsInferred)",
			specJSON,
			inferredJSON,
			varsJSON,
		)),
		jpoet.WriterOutput(&out),
	)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
