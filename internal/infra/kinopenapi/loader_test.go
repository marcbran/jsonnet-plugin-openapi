package kinopenapi

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestPruneNonGETPaths(t *testing.T) {
	tests := []struct {
		name     string
		paths    *openapi3.Paths
		wantKeys []string
	}{
		{
			name:     "nil paths",
			paths:    nil,
			wantKeys: nil,
		},
		{
			name: "keeps paths with a GET operation",
			paths: openapi3.NewPaths(
				openapi3.WithPath("/x", &openapi3.PathItem{Get: &openapi3.Operation{}}),
			),
			wantKeys: []string{"/x"},
		},
		{
			name: "drops paths with no GET operation",
			paths: openapi3.NewPaths(
				openapi3.WithPath("/get", &openapi3.PathItem{Get: &openapi3.Operation{}}),
				openapi3.WithPath("/delete", &openapi3.PathItem{Delete: &openapi3.Operation{}}),
			),
			wantKeys: []string{"/get"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &openapi3.T{Paths: tt.paths}
			pruneNonGETPaths(doc)
			var gotKeys []string
			if doc.Paths != nil {
				gotKeys = doc.Paths.Keys()
			}
			require.ElementsMatch(t, tt.wantKeys, gotKeys)
		})
	}
}
