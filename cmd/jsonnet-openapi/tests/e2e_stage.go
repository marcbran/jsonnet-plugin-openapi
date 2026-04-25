//go:build e2e

package tests

import (
	"testing"

	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	"github.com/stretchr/testify/require"
)

type Stage struct {
	t require.TestingT

	facade openapipkg.Facade

	tempDir  string
	outDir   string
	ref      string
	service  string

	lastOutput openapipkg.Output
	lastErr    string

	liveHTTPOrigin string

	evalOut map[string]any
	evalErr error

	batchJobs        []openapipkg.Input
	lastBatchOutputs []openapipkg.Output
	lastBatchErr     string
}

func scenario(t *testing.T) (*Stage, *Stage, *Stage) {
	facade, err := NewCLIFacade()
	require.NoError(t, err)
	tempDir := t.TempDir()
	s := &Stage{
		t:       t,
		facade:  facade,
		tempDir: tempDir,
		outDir:  tempDir,
	}
	return s, s, s
}

func (s *Stage) and() *Stage {
	return s
}
