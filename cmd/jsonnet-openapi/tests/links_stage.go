//go:build e2e

package tests

import (
	"context"
	"os"
	"path/filepath"

	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	"github.com/stretchr/testify/require"
)

func (s *Stage) a_links_spec_file(name string) *Stage {
	s.linksSpec = filepath.Join(testdataRoot(), name)
	return s
}

func (s *Stage) a_links_output_under_temp(name string) *Stage {
	s.linksOut = filepath.Join(s.tempDir, name)
	return s
}

func (s *Stage) a_links_workdir_under_temp(name string) *Stage {
	s.linksWorkDir = filepath.Join(s.tempDir, name)
	return s
}

func (s *Stage) writeLinksCache(jobName, taskID, content string) *Stage {
	path := filepath.Join(s.linksWorkDir, jobName, "results", taskID+".json")
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(s.t, err)
	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(s.t, err)
	return s
}

func (s *Stage) a_cached_account_link_from_resource() *Stage {
	return s.writeLinksCache(
		"links-from-resources",
		"accounts___id",
		`{"links":[{"sourcePath":"/accounts/{id}","at":["members"],"targetPath":"/users/{userId}","keys":[{"const":"members","path":null},{"const":null,"path":["userId"]}],"confidence":"high","reason":"members items reference a user via userId, which matches the getUser detail endpoint."}]}`,
	)
}

func (s *Stage) a_cached_user_link_from_resource() *Stage {
	return s.writeLinksCache(
		"links-from-resources",
		"users___userId",
		`{"links":[]}`,
	)
}

func (s *Stage) a_cached_users_link_from_collection() *Stage {
	return s.writeLinksCache(
		"links-from-collections",
		"users",
		`{"links":[{"sourcePath":"/users","at":[],"targetPath":"/users/{userId}","keys":[{"const":null,"path":["id"]}],"confidence":"high","reason":"The list returns User records whose canonical detail GET is getUser."}]}`,
	)
}

// /reports wraps its item array in allOf with a pagination schema and inlines
// the item schema, so it is only recognised as a collection by the structural
// rule -- a $ref-based or top-level-properties check misses it entirely.
func (s *Stage) a_cached_reports_link_from_collection() *Stage {
	return s.writeLinksCache(
		"links-from-collections",
		"reports",
		`{"links":[]}`,
	)
}

func (s *Stage) a_cached_member_link_vars() *Stage {
	return s.writeLinksCache(
		"link-vars",
		"accounts___id--users___userId--at_members--k_members__p_userId",
		`{"sourcePath":"/accounts/{id}","targetPath":"/users/{userId}","at":["members"],"keys":[{"const":"members","path":null},{"const":null,"path":["userId"]}],"vars":[{"param":"userId","path":["userId"]}],"confidence":"high","reason":"Each member object carries its own userId field."}`,
	)
}

func (s *Stage) a_cached_users_link_vars() *Stage {
	return s.writeLinksCache(
		"link-vars",
		"users--users___userId--k_p_id",
		`{"sourcePath":"/users","targetPath":"/users/{userId}","at":[],"keys":[{"const":null,"path":["id"]}],"vars":[{"param":"userId","path":["id"]}],"confidence":"high","reason":"The list item schema exposes id as the stable user identifier."}`,
	)
}

func (s *Stage) the_links_command_is_run() *Stage {
	out, err := s.facade.InferLinks(context.Background(), openapipkg.LinksInput{
		Spec:    s.linksSpec,
		Out:     s.linksOut,
		WorkDir: s.linksWorkDir,
	})
	if err != nil {
		s.lastLinksOutput = out
		s.lastLinksErr = err.Error()
		return s
	}
	s.lastLinksOutput = out
	s.lastLinksErr = ""
	return s
}

func (s *Stage) the_links_has_no_error() *Stage {
	require.Empty(s.t, s.lastLinksErr)
	return s
}

func (s *Stage) the_links_file_matches(fixture string) *Stage {
	raw, err := os.ReadFile(s.lastLinksOutput.Out)
	require.NoError(s.t, err)
	expected, err := os.ReadFile(filepath.Join(testdataRoot(), fixture))
	require.NoError(s.t, err)
	require.Equal(s.t, string(expected), string(raw))
	return s
}

func (s *Stage) the_links_output_path_is_under_temp(name string) *Stage {
	require.Equal(s.t, filepath.Join(s.tempDir, name), s.lastLinksOutput.Out)
	return s
}
