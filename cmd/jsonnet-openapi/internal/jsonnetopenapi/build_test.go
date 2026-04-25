package jsonnetopenapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildPayload(t *testing.T) {
	leaf := func(id, pathTmpl, pathFmt string, pargs []string, qp, hp []ParamSpec) *GenOperation {
		if pargs == nil {
			pargs = []string{}
		}
		if qp == nil {
			qp = []ParamSpec{}
		}
		if hp == nil {
			hp = []ParamSpec{}
		}
		return &GenOperation{
			ID: id, PathTemplate: pathTmpl, PathFormat: pathFmt,
			PathArgNames: pargs, QueryParams: qp, HeaderParams: hp,
		}
	}
	trieLeaf := func(op *GenOperation) *TrieNode {
		return &TrieNode{Leaf: op, Children: map[string]*TrieNode{}}
	}
	tests := []struct {
		name        string
		api         APISpec
		override    string
		hint        string
		wantErr     string
		wantPayload *GenPayload
	}{
		{
			name:    "no GET operations",
			api:     APISpec{Title: "T", Version: "1"},
			wantErr: "no GET operations in spec",
		},
		{
			name: "duplicate GET for same path",
			api: APISpec{
				Paths: []PathItem{
					{Path: "/x", Get: &Operation{OperationID: "a"}},
					{Path: "/x", Get: &Operation{OperationID: "b"}},
				},
			},
			wantErr: "duplicate GET",
		},
		{
			name:     "service override wins over title",
			api:      APISpec{Title: "Nice Title", Paths: []PathItem{{Path: "/p", Get: &Operation{OperationID: "op"}}}},
			override: "override-slug",
			wantPayload: &GenPayload{
				Info:    GenInfo{Title: "Nice Title", Version: ""},
				Service: "override-slug",
				Trie: &TrieNode{
					Children: map[string]*TrieNode{
						"p": trieLeaf(leaf("op", "/p", "/p", []string{}, nil, nil)),
					},
				},
			},
		},
		{
			name: "service slugified from title",
			api:  APISpec{Title: "My API", Paths: []PathItem{{Path: "/a", Get: &Operation{OperationID: "x"}}}},
			wantPayload: &GenPayload{
				Info:    GenInfo{Title: "My API", Version: ""},
				Service: "my-api",
				Trie: &TrieNode{
					Children: map[string]*TrieNode{
						"a": trieLeaf(leaf("x", "/a", "/a", []string{}, nil, nil)),
					},
				},
			},
		},
		{
			name: "service from file path hint when title empty",
			api:  APISpec{Paths: []PathItem{{Path: "/a", Get: &Operation{OperationID: "x"}}}},
			hint: "/var/app/openapi.yaml",
			wantPayload: &GenPayload{
				Info:    GenInfo{Title: "", Version: ""},
				Service: "openapi",
				Trie: &TrieNode{
					Children: map[string]*TrieNode{
						"a": trieLeaf(leaf("x", "/a", "/a", []string{}, nil, nil)),
					},
				},
			},
		},
		{
			name: "service from URL path hint when title empty",
			api:  APISpec{Paths: []PathItem{{Path: "/a", Get: &Operation{OperationID: "x"}}}},
			hint: "https://example.com/v1/spec.json",
			wantPayload: &GenPayload{
				Info:    GenInfo{Title: "", Version: ""},
				Service: "spec",
				Trie: &TrieNode{
					Children: map[string]*TrieNode{
						"a": trieLeaf(leaf("x", "/a", "/a", []string{}, nil, nil)),
					},
				},
			},
		},
		{
			name:    "cannot derive service without title path or override",
			api:     APISpec{Paths: []PathItem{{Path: "/", Get: &Operation{OperationID: "root"}}}},
			wantErr: "cannot derive service",
		},
		{
			name:     "service override unusable after sanitization",
			api:      APISpec{Title: "T", Paths: []PathItem{{Path: "/p", Get: &Operation{OperationID: "op"}}}},
			override: "@@@",
			wantErr:  "sanitization",
		},
		{
			name: "service from title starting with digit is prefixed",
			api: APISpec{
				Title: "123-api",
				Paths: []PathItem{{Path: "/a", Get: &Operation{OperationID: "x"}}},
			},
			wantPayload: &GenPayload{
				Info:    GenInfo{Title: "123-api", Version: ""},
				Service: "s-123-api",
				Trie: &TrieNode{
					Children: map[string]*TrieNode{
						"a": trieLeaf(leaf("x", "/a", "/a", []string{}, nil, nil)),
					},
				},
			},
		},
		{
			name: "info version passed through",
			api: APISpec{
				Title: "T", Version: "2.0",
				Paths: []PathItem{{Path: "/z", Get: &Operation{OperationID: "z"}}},
			},
			override: "svc",
			wantPayload: &GenPayload{
				Info:    GenInfo{Title: "T", Version: "2.0"},
				Service: "svc",
				Trie: &TrieNode{
					Children: map[string]*TrieNode{
						"z": trieLeaf(leaf("z", "/z", "/z", []string{}, nil, nil)),
					},
				},
			},
		},
		{
			name: "operation id derived from path when missing",
			api: APISpec{
				Paths: []PathItem{{Path: "/foo/bar", Get: &Operation{OperationID: ""}}},
			},
			override: "s",
			wantPayload: &GenPayload{
				Info:    GenInfo{Title: "", Version: ""},
				Service: "s",
				Trie: &TrieNode{
					Children: map[string]*TrieNode{
						"foo": {
							Children: map[string]*TrieNode{
								"bar": trieLeaf(leaf("get_foo-bar", "/foo/bar", "/foo/bar", []string{}, nil, nil)),
							},
						},
					},
				},
			},
		},
		{
			name: "query and header parameters",
			api: APISpec{
				Paths: []PathItem{{
					Path: "/q",
					Get: &Operation{Parameters: []Parameter{
						{Name: "q", In: "query", Required: false},
						{Name: "X-Trace", In: "header", Required: true},
					}},
				}},
			},
			override: "s",
			wantPayload: &GenPayload{
				Info:    GenInfo{Title: "", Version: ""},
				Service: "s",
				Trie: &TrieNode{
					Children: map[string]*TrieNode{
						"q": trieLeaf(&GenOperation{
							ID:           "get_q",
							PathTemplate: "/q",
							PathFormat:   "/q",
							PathArgNames: []string{},
							QueryParams:  []ParamSpec{{Name: "q", Required: false}},
							HeaderParams: []ParamSpec{{Name: "X-Trace", Required: true}},
						}),
					},
				},
			},
		},
		{
			name: "two GET routes with shared prefix promote static leaf to underscore",
			api: APISpec{
				Paths: []PathItem{
					{Path: "/r", Get: &Operation{OperationID: "one"}},
					{
						Path: "/r/{id}",
						Get: &Operation{
							OperationID: "two",
							Parameters:  []Parameter{{Name: "id", In: "path", Required: true}},
						},
					},
				},
			},
			override: "mysvc",
			wantPayload: &GenPayload{
				Info:    GenInfo{Title: "", Version: ""},
				Service: "mysvc",
				Trie: &TrieNode{
					Children: map[string]*TrieNode{
						"r": {
							Children: map[string]*TrieNode{
								"_": {Leaf: leaf("one", "/r", "/r", []string{}, nil, nil)},
								"{id}": trieLeaf(leaf("two", "/r/{id}", "/r/%s", []string{"id"}, nil, nil)),
							},
						},
					},
				},
			},
		},
		{
			name: "operation parameters override path parameters during build",
			api: APISpec{
				Paths: []PathItem{{
					Path: "/q",
					Parameters: []Parameter{
						{Name: "q", In: "query", Required: false},
					},
					Get: &Operation{
						Parameters: []Parameter{
							{Name: "q", In: "query", Required: true},
						},
					},
				}},
			},
			override: "s",
			wantPayload: &GenPayload{
				Info:    GenInfo{Title: "", Version: ""},
				Service: "s",
				Trie: &TrieNode{
					Children: map[string]*TrieNode{
						"q": trieLeaf(&GenOperation{
							ID:           "get_q",
							PathTemplate: "/q",
							PathFormat:   "/q",
							PathArgNames: []string{},
							QueryParams:  []ParamSpec{{Name: "q", Required: true}},
							HeaderParams: []ParamSpec{},
						}),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildPayload(tt.api, tt.override, tt.hint)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantPayload, got)
		})
	}
}

func TestSanitizeServiceSlug(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "letters", in: "My API", want: "my-api"},
		{name: "leading digit", in: "123x", want: "s-123x"},
		{name: "empty", in: "", want: ""},
		{name: "no letters", in: "@@@", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, sanitizeServiceSlug(tt.in))
		})
	}
}

func TestCanonicalizeParameters(t *testing.T) {
	tests := []struct {
		name string
		in   []Parameter
		want []Parameter
	}{
		{
			name: "empty",
			in:   []Parameter{},
			want: []Parameter{},
		},
		{
			name: "single",
			in:   []Parameter{{Name: "a", In: "query"}},
			want: []Parameter{{Name: "a", In: "query"}},
		},
		{
			name: "sorts by in then name",
			in: []Parameter{
				{Name: "zebra", In: "query"},
				{Name: "HeaderB", In: "header"},
				{Name: "alpha", In: "query"},
				{Name: "HeaderA", In: "header"},
			},
			want: []Parameter{
				{Name: "HeaderA", In: "header"},
				{Name: "HeaderB", In: "header"},
				{Name: "alpha", In: "query"},
				{Name: "zebra", In: "query"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canonicalizeParameters(tt.in)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPathFormatFromTemplate(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantFormat string
		wantNames  []string
	}{
		{name: "single path param", path: "/users/{userId}", wantFormat: "/users/%s", wantNames: []string{"userId"}},
		{name: "no params", path: "/users", wantFormat: "/users", wantNames: []string{}},
		{name: "root", path: "/", wantFormat: "/", wantNames: []string{}},
		{name: "two params", path: "/a/{x}/b/{y}", wantFormat: "/a/%s/b/%s", wantNames: []string{"x", "y"}},
		{name: "wildcard suffix stripped", path: "/v/{name*}", wantFormat: "/v/%s", wantNames: []string{"name"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, n := pathFormatFromTemplate(tt.path)
			require.Equal(t, tt.wantFormat, f)
			require.Equal(t, tt.wantNames, n)
		})
	}
}
