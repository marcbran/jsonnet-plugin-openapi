//go:build e2e

package tests

import (
	"context"
	"os"
	"path/filepath"

	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	"github.com/stretchr/testify/require"
)

func (s *Stage) a_list_columns_spec_file(name string) *Stage {
	s.listColumnsSpec = filepath.Join(testdataRoot(), name)
	return s
}

func (s *Stage) a_list_columns_output_under_temp(name string) *Stage {
	s.listColumnsOut = filepath.Join(s.tempDir, name)
	return s
}

func (s *Stage) a_list_columns_workdir_under_temp(name string) *Stage {
	s.listColumnsWorkDir = filepath.Join(s.tempDir, name)
	return s
}

func (s *Stage) a_cached_user_columns_inference() *Stage {
	path := filepath.Join(
		s.listColumnsWorkDir,
		"list-columns-inference",
		"results",
		"users.json",
	)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(s.t, err)
	err = os.WriteFile(path, []byte(`{"sourcePath":"/users","operationId":"listUsers","array":[],"columns":[{"key":"id","path":["id"],"label":"ID","kind":"identifier","priority":"primary","reason":"Stable user identifier."},{"key":"name","path":["name"],"label":"Name","kind":"name","priority":"secondary","reason":"Human-readable user name."},{"key":"status","path":["status"],"label":"Status","kind":"status","priority":"secondary","reason":"Useful account state."},{"key":"createdAt","path":["createdAt"],"label":"Created","kind":"timestamp","priority":"tertiary","reason":"Useful recency signal."}],"confidence":"high","reason":"The schema exposes common scalar fields for a user table."}`), 0644)
	require.NoError(s.t, err)
	return s
}

func (s *Stage) the_list_columns_command_is_run() *Stage {
	out, err := s.facade.InferListColumns(context.Background(), openapipkg.ListColumnsInput{
		Spec:    s.listColumnsSpec,
		Out:     s.listColumnsOut,
		WorkDir: s.listColumnsWorkDir,
	})
	if err != nil {
		s.lastListColumnsOutput = out
		s.lastListColumnsErr = err.Error()
		return s
	}
	s.lastListColumnsOutput = out
	s.lastListColumnsErr = ""
	return s
}

func (s *Stage) the_list_columns_has_no_error() *Stage {
	require.Empty(s.t, s.lastListColumnsErr)
	return s
}

func (s *Stage) the_columns_file_matches(fixture string) *Stage {
	raw, err := os.ReadFile(s.lastListColumnsOutput.Out)
	require.NoError(s.t, err)
	expected, err := os.ReadFile(filepath.Join(testdataRoot(), fixture))
	require.NoError(s.t, err)
	require.Equal(s.t, string(expected), string(raw))
	return s
}

func (s *Stage) the_columns_output_path_is_under_temp(name string) *Stage {
	require.Equal(s.t, filepath.Join(s.tempDir, name), s.lastListColumnsOutput.Out)
	return s
}
