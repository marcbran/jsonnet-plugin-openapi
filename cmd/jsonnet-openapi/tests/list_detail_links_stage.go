//go:build e2e

package tests

import (
	"context"
	"os"
	"path/filepath"

	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	"github.com/stretchr/testify/require"
)

func (s *Stage) a_list_detail_links_spec(name string) *Stage {
	s.listDetailLinksSpec = filepath.Join(testdataRoot(), name+".json")
	return s
}

func (s *Stage) a_list_detail_links_spec_file(name string) *Stage {
	s.listDetailLinksSpec = filepath.Join(testdataRoot(), name)
	return s
}

func (s *Stage) a_list_detail_links_output_under_temp(name string) *Stage {
	s.listDetailLinksOut = filepath.Join(s.tempDir, name)
	return s
}

func (s *Stage) a_list_detail_links_workdir_under_temp(name string) *Stage {
	s.listDetailLinksWorkDir = filepath.Join(s.tempDir, name)
	return s
}

func (s *Stage) a_cached_user_detail_var_inference() *Stage {
	path := filepath.Join(
		s.listDetailLinksWorkDir,
		"list-detail-vars-inference",
		"results",
		"users--users___userId.json",
	)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(s.t, err)
	err = os.WriteFile(path, []byte(`{"sourcePath":"/users","targetPath":"/users/{userId}","array":[],"vars":[{"param":"userId","path":["id"]}],"confidence":"high","reason":"The list item schema exposes id as the stable user identifier."}`), 0644)
	require.NoError(s.t, err)
	return s
}

func (s *Stage) the_list_detail_links_command_is_run() *Stage {
	out, err := s.facade.InferListDetailLinks(context.Background(), openapipkg.ListDetailLinksInput{
		Spec:    s.listDetailLinksSpec,
		Out:     s.listDetailLinksOut,
		WorkDir: s.listDetailLinksWorkDir,
	})
	if err != nil {
		s.lastListDetailLinksOutput = out
		s.lastListDetailLinksErr = err.Error()
		return s
	}
	s.lastListDetailLinksOutput = out
	s.lastListDetailLinksErr = ""
	return s
}

func (s *Stage) the_list_detail_links_has_no_error() *Stage {
	require.Empty(s.t, s.lastListDetailLinksErr)
	return s
}

func (s *Stage) the_links_file_matches(fixture string) *Stage {
	raw, err := os.ReadFile(s.lastListDetailLinksOutput.Out)
	require.NoError(s.t, err)
	expected, err := os.ReadFile(filepath.Join(testdataRoot(), fixture))
	require.NoError(s.t, err)
	require.Equal(s.t, string(expected), string(raw))
	return s
}

func (s *Stage) the_links_output_path_is_under_temp(name string) *Stage {
	require.Equal(s.t, filepath.Join(s.tempDir, name), s.lastListDetailLinksOutput.Out)
	return s
}
