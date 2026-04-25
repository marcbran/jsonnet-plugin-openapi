//go:build e2e

package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/marcbran/jpoet/pkg/jpoet"
	openapiplugin "github.com/marcbran/jsonnet-plugin-openapi/openapi"
	"github.com/stretchr/testify/require"
)

type SpecStage struct {
	t    require.TestingT
	spec string
	out  map[string]any
	err  error
}

func specScenario(t *testing.T) (*SpecStage, *SpecStage, *SpecStage) {
	s := &SpecStage{t: t}
	return s, s, s
}

func (s *SpecStage) and() *SpecStage {
	return s
}

func (s *SpecStage) a_fixture_spec(name string) *SpecStage {
	_, file, _, ok := runtime.Caller(0)
	require.True(s.t, ok)
	root := filepath.Join(filepath.Dir(file), "..", "cmd", "jsonnet-openapi", "tests", "testdata")
	raw, err := os.ReadFile(filepath.Join(root, name+".yaml"))
	require.NoError(s.t, err)
	s.spec = string(raw)
	return s
}

func (s *SpecStage) the_jsonnet_api_spec_is_evaluated() *SpecStage {
	s.err = jpoet.Eval(
		jpoet.WithPlugin(openapiplugin.Plugin()),
		jpoet.SnippetInput("test.jsonnet", fmt.Sprintf(`std.native('invoke:openapi')('apiSpec', [%q])`, s.spec)),
		jpoet.ValueOutput(&s.out),
		jpoet.Serialize(false),
	)
	return s
}

func (s *SpecStage) the_jsonnet_nested_spec_is_evaluated() *SpecStage {
	s.err = jpoet.Eval(
		jpoet.WithPlugin(openapiplugin.Plugin()),
		jpoet.SnippetInput("test.jsonnet", fmt.Sprintf(`std.native('invoke:openapi')('nestedSpec', [%q])`, s.spec)),
		jpoet.ValueOutput(&s.out),
		jpoet.Serialize(false),
	)
	return s
}

func (s *SpecStage) the_eval_has_no_error() *SpecStage {
	require.NoError(s.t, s.err)
	return s
}

func (s *SpecStage) the_result_matches_expected_file(fixture string, name string) *SpecStage {
	_, file, _, ok := runtime.Caller(0)
	require.True(s.t, ok)
	root := filepath.Join(filepath.Dir(file), "..", "cmd", "jsonnet-openapi", "tests", "testdata", fixture)

	wantRaw, err := os.ReadFile(filepath.Join(root, name))
	require.NoError(s.t, err)

	var want any
	err = json.Unmarshal(wantRaw, &want)
	require.NoError(s.t, err)

	gotJSON, err := marshalCanonicalJSON(s.out)
	require.NoError(s.t, err)
	wantJSON, err := marshalCanonicalJSON(want)
	require.NoError(s.t, err)

	require.Equal(s.t, string(wantJSON), string(gotJSON))
	return s
}

func marshalCanonicalJSON(v any) ([]byte, error) {
	return json.MarshalIndent(normalizeJSON(v), "", "  ")
}

func normalizeJSON(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			out[k] = normalizeJSON(vv)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = normalizeJSON(t[i])
		}
		if isPathItemArray(out) {
			sort.Slice(out, func(i, j int) bool {
				li := out[i].(map[string]any)["path"].(string)
				lj := out[j].(map[string]any)["path"].(string)
				return li < lj
			})
		}
		return out
	default:
		return v
	}
}

func isPathItemArray(v []any) bool {
	if len(v) == 0 {
		return false
	}
	for _, item := range v {
		m, ok := item.(map[string]any)
		if !ok {
			return false
		}
		if _, ok := m["path"].(string); !ok {
			return false
		}
	}
	return true
}
