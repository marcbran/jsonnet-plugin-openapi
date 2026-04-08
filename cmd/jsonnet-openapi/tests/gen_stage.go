//go:build e2e

package tests

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	"github.com/google/go-jsonnet/formatter"
	"github.com/stretchr/testify/require"
)

func testdataRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "testdata")
}

func (s *Stage) a_spec(name string) *Stage {
	s.specPath = filepath.Join(testdataRoot(), name+".yaml")
	return s
}

func (s *Stage) a_service(name string) *Stage {
	s.service = name
	return s
}

func (s *Stage) the_gen_command_is_run() *Stage {
	out, err := s.facade.Generate(context.Background(), openapipkg.Input{
		Spec:    s.specPath,
		OutDir:  s.outDir,
		Service: s.service,
		PkgRepo: "git@github.com:marcbran/jsonnet.git",
	})
	if err != nil {
		s.lastOutput = out
		s.lastErr = err.Error()
		return s
	}
	s.lastOutput = out
	s.lastErr = ""
	return s
}

func (s *Stage) the_gen_has_no_error() *Stage {
	require.Empty(s.t, s.lastErr)
	return s
}

func (s *Stage) the_generated_files_match(name string) *Stage {
	expectedDir := filepath.Join(testdataRoot(), name)
	names := []string{"openapi.resolved.json", "main.libsonnet", "pkg.libsonnet"}
	for _, fname := range names {
		gotPath := filepath.Join(s.outDir, fname)
		wantPath := filepath.Join(expectedDir, fname)
		got, err := os.ReadFile(gotPath)
		require.NoError(s.t, err)
		want, err := os.ReadFile(wantPath)
		require.NoError(s.t, err)
		require.Equal(s.t, string(want), string(got))
	}
	return s
}

func (s *Stage) the_generated_main_libsonnet_parses_as_jsonnet() *Stage {
	src, err := os.ReadFile(filepath.Join(s.outDir, "main.libsonnet"))
	require.NoError(s.t, err)
	_, _, err = formatter.SnippetToRawAST("main.libsonnet", string(src))
	require.NoError(s.t, err)
	return s
}
