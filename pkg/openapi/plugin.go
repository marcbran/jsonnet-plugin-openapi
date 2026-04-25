package openapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/marcbran/jpoet/pkg/jpoet"
	kinopenapi "github.com/marcbran/jsonnet-plugin-openapi/internal/infra/kinopenapi"
	internalopenapi "github.com/marcbran/jsonnet-plugin-openapi/internal/openapi"
)

func Plugin() *jpoet.Plugin {
	parser := kinopenapi.NewLoader()
	return jpoet.NewPlugin("openapi", []jsonnet.NativeFunction{
		APISpec(parser),
		NestedSpec(parser),
	})
}

func APISpec(parser internalopenapi.Parser) jsonnet.NativeFunction {
	return jsonnet.NativeFunction{
		Name:   "apiSpec",
		Params: ast.Identifiers{"spec"},
		Func: func(input []any) (any, error) {
			spec, err := parseSpecInput(input)
			if err != nil {
				return nil, err
			}
			api, err := parser.Parse(context.Background(), spec)
			if err != nil {
				return nil, err
			}
			return toJSONValue(api)
		},
	}
}

func NestedSpec(parser internalopenapi.Parser) jsonnet.NativeFunction {
	return jsonnet.NativeFunction{
		Name:   "nestedSpec",
		Params: ast.Identifiers{"spec"},
		Func: func(input []any) (any, error) {
			spec, err := parseSpecInput(input)
			if err != nil {
				return nil, err
			}
			api, err := parser.Parse(context.Background(), spec)
			if err != nil {
				return nil, err
			}
			nested, err := internalopenapi.BuildNestedSpec(api)
			if err != nil {
				return nil, err
			}
			return toJSONValue(nested)
		},
	}
}

func parseSpecInput(input []any) (string, error) {
	if len(input) != 1 {
		return "", fmt.Errorf("expected spec")
	}
	spec, ok := input[0].(string)
	if !ok || spec == "" {
		return "", fmt.Errorf("spec must be a non-empty string")
	}
	return spec, nil
}

func toJSONValue(v any) (any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
