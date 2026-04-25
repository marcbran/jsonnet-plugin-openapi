//go:build e2e

package tests

import (
	"context"
	"path/filepath"

	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	"github.com/stretchr/testify/require"
)

func (s *Stage) a_batch_job_from_testdata(specBaseName string, outputDirUnderTemp string, service string) *Stage {
	s.batchJobs = append(s.batchJobs, openapipkg.Input{
		Ref:     filepath.Join(testdataRoot(), specBaseName+".yaml"),
		OutDir:  filepath.Join(s.tempDir, outputDirUnderTemp),
		Service: service,
	})
	return s
}

func (s *Stage) a_batch_job_with_missing_spec_file(outputDirUnderTemp string, service string) *Stage {
	s.batchJobs = append(s.batchJobs, openapipkg.Input{
		Ref:     filepath.Join(s.tempDir, "this-spec-file-does-not-exist.yaml"),
		OutDir:  filepath.Join(s.tempDir, outputDirUnderTemp),
		Service: service,
	})
	return s
}

func (s *Stage) the_batch_command_is_run() *Stage {
	outs, err := s.facade.Batch(context.Background(), s.batchJobs)
	if err != nil {
		s.lastBatchErr = err.Error()
		s.lastBatchOutputs = nil
		return s
	}
	s.lastBatchErr = ""
	s.lastBatchOutputs = outs
	return s
}

func (s *Stage) the_batch_has_no_error() *Stage {
	require.Empty(s.t, s.lastBatchErr)
	return s
}

func (s *Stage) the_batch_has_an_error() *Stage {
	require.NotEmpty(s.t, s.lastBatchErr)
	return s
}

func (s *Stage) the_batch_outputs_are_nil() *Stage {
	require.Nil(s.t, s.lastBatchOutputs)
	return s
}

func (s *Stage) the_batch_job_count_is(n int) *Stage {
	require.Len(s.t, s.lastBatchOutputs, n)
	return s
}

func (s *Stage) the_batch_output_out_dirs_under_temp_are(relPaths ...string) *Stage {
	require.Len(s.t, s.lastBatchOutputs, len(relPaths))
	for i, rel := range relPaths {
		want := filepath.Join(s.tempDir, rel)
		require.Equal(s.t, want, s.lastBatchOutputs[i].OutDir)
	}
	return s
}

func (s *Stage) the_generated_main_libsonnet_exists_for_batch_job(index int) *Stage {
	require.Greater(s.t, len(s.lastBatchOutputs), index)
	outDir := s.lastBatchOutputs[index].OutDir
	require.FileExists(s.t, filepath.Join(outDir, "main.libsonnet"))
	return s
}

func (s *Stage) the_generated_main_libsonnet_exists_under_temp(relPath string) *Stage {
	require.FileExists(s.t, filepath.Join(s.tempDir, relPath, "main.libsonnet"))
	return s
}

func (s *Stage) the_generated_main_libsonnet_does_not_exist_under_temp(relPath string) *Stage {
	require.NoFileExists(s.t, filepath.Join(s.tempDir, relPath, "main.libsonnet"))
	return s
}
