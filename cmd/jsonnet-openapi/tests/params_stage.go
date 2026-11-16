//go:build e2e

package tests

import (
	"context"
	"os"
	"path/filepath"

	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	"github.com/stretchr/testify/require"
)

func (s *Stage) a_params_spec_file(name string) *Stage {
	s.paramsSpec = filepath.Join(testdataRoot(), name)
	return s
}

func (s *Stage) a_params_output_under_temp(name string) *Stage {
	s.paramsOut = filepath.Join(s.tempDir, name)
	return s
}

func (s *Stage) a_params_workdir_under_temp(name string) *Stage {
	s.paramsWorkDir = filepath.Join(s.tempDir, name)
	return s
}

func (s *Stage) a_cached_user_params_inference() *Stage {
	path := filepath.Join(
		s.paramsWorkDir,
		"params-inference",
		"results",
		"users.json",
	)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(s.t, err)
	err = os.WriteFile(path, []byte(`{"sourcePath":"/users","operationId":"listUsers","params":[{"name":"include[]","role":"shaping","reason":"Sideloads related teams."},{"name":"limit","role":"pagination","reason":"Page size."},{"name":"sort","role":"sort","reason":"Sort key."},{"name":"status","role":"filter","reason":"Narrows users by status."}],"confidence":"high","reason":"Each param has a clear purpose."}`), 0644)
	require.NoError(s.t, err)
	return s
}

func (s *Stage) the_params_command_is_run() *Stage {
	out, err := s.facade.InferParams(context.Background(), openapipkg.ParamsInput{
		Spec:    s.paramsSpec,
		Out:     s.paramsOut,
		WorkDir: s.paramsWorkDir,
	})
	if err != nil {
		s.lastParamsOutput = out
		s.lastParamsErr = err.Error()
		return s
	}
	s.lastParamsOutput = out
	s.lastParamsErr = ""
	return s
}

func (s *Stage) the_params_has_no_error() *Stage {
	require.Empty(s.t, s.lastParamsErr)
	return s
}

func (s *Stage) the_params_file_matches(fixture string) *Stage {
	raw, err := os.ReadFile(s.lastParamsOutput.Out)
	require.NoError(s.t, err)
	expected, err := os.ReadFile(filepath.Join(testdataRoot(), fixture))
	require.NoError(s.t, err)
	require.Equal(s.t, string(expected), string(raw))
	return s
}

func (s *Stage) the_params_output_path_is_under_temp(name string) *Stage {
	require.Equal(s.t, filepath.Join(s.tempDir, name), s.lastParamsOutput.Out)
	return s
}
