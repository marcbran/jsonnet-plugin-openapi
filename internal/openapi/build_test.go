package openapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildNestedSpec(t *testing.T) {
	leaf := func(id, pathTmpl, pathFmt string, pargs []string, qp, hp []ParameterSpec) *OperationSpec {
		if pargs == nil {
			pargs = []string{}
		}
		if qp == nil {
			qp = []ParameterSpec{}
		}
		if hp == nil {
			hp = []ParameterSpec{}
		}
		return &OperationSpec{
			ID: id, PathTemplate: pathTmpl, PathFormat: pathFmt,
			PathArgNames: pargs, QueryParams: qp, HeaderParams: hp,
		}
	}
	pathLeaf := func(op *OperationSpec) *PathNode {
		return &PathNode{Operation: op, Children: map[string]*PathNode{}}
	}
	tests := []struct {
		name     string
		api      APISpec
		wantErr  string
		wantSpec *NestedSpec
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
			name: "title and version passed through",
			api:  APISpec{Title: "Nice Title", Paths: []PathItem{{Path: "/p", Get: &Operation{OperationID: "op"}}}},
			wantSpec: &NestedSpec{
				Title:   "Nice Title",
				Version: "",
				Paths: &PathNode{
					Children: map[string]*PathNode{
						"p": pathLeaf(leaf("op", "/p", "/p", []string{}, nil, nil)),
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
			wantSpec: &NestedSpec{
				Title:   "T",
				Version: "2.0",
				Paths: &PathNode{
					Children: map[string]*PathNode{
						"z": pathLeaf(leaf("z", "/z", "/z", []string{}, nil, nil)),
					},
				},
			},
		},
		{
			name: "operation id derived from path when missing",
			api: APISpec{
				Paths: []PathItem{{Path: "/foo/bar", Get: &Operation{OperationID: ""}}},
			},
			wantSpec: &NestedSpec{
				Title:   "",
				Version: "",
				Paths: &PathNode{
					Children: map[string]*PathNode{
						"foo": {
							Children: map[string]*PathNode{
								"bar": pathLeaf(leaf("get_foo-bar", "/foo/bar", "/foo/bar", []string{}, nil, nil)),
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
			wantSpec: &NestedSpec{
				Title:   "",
				Version: "",
				Paths: &PathNode{
					Children: map[string]*PathNode{
						"q": pathLeaf(&OperationSpec{
							ID:           "get_q",
							PathTemplate: "/q",
							PathFormat:   "/q",
							PathArgNames: []string{},
							QueryParams:  []ParameterSpec{{Name: "q", Required: false}},
							HeaderParams: []ParameterSpec{{Name: "X-Trace", Required: true}},
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
			wantSpec: &NestedSpec{
				Title:   "",
				Version: "",
				Paths: &PathNode{
					Children: map[string]*PathNode{
						"r": {
							Children: map[string]*PathNode{
								"_":    {Operation: leaf("one", "/r", "/r", []string{}, nil, nil)},
								"{id}": pathLeaf(leaf("two", "/r/{id}", "/r/%s", []string{"id"}, nil, nil)),
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
			wantSpec: &NestedSpec{
				Title:   "",
				Version: "",
				Paths: &PathNode{
					Children: map[string]*PathNode{
						"q": pathLeaf(&OperationSpec{
							ID:           "get_q",
							PathTemplate: "/q",
							PathFormat:   "/q",
							PathArgNames: []string{},
							QueryParams:  []ParameterSpec{{Name: "q", Required: true}},
							HeaderParams: []ParameterSpec{},
						}),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildNestedSpec(tt.api)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantSpec, got)
		})
	}
}

func TestResolveServiceName(t *testing.T) {
	tests := []struct {
		name     string
		override string
		title    string
		hint     string
		want     string
		wantErr  string
	}{
		{name: "service override wins over title", override: "override-slug", title: "Nice Title", want: "override-slug"},
		{name: "service slugified from title", title: "My API", want: "my-api"},
		{name: "service from file path hint when title empty", hint: "/var/app/openapi.yaml", want: "openapi"},
		{name: "service from URL path hint when title empty", hint: "https://example.com/v1/spec.json", want: "spec"},
		{name: "cannot derive service without title path or override", wantErr: "cannot derive service"},
		{name: "service override unusable after sanitization", override: "@@@", title: "T", wantErr: "sanitization"},
		{name: "service from title starting with digit is prefixed", title: "123-api", want: "s-123-api"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveServiceName(tt.override, tt.title, tt.hint)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
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
