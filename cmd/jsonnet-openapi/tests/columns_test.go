//go:build e2e

package tests

import "testing"

func TestColumnsInferWithCachedInference(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_columns_spec_file("columns.yaml").and().
		a_columns_output_under_temp("columns.columns.json").and().
		a_columns_workdir_under_temp("columns-work").and().
		a_cached_user_columns_inference()

	when.
		the_columns_command_is_run()

	then.
		the_columns_has_no_error().and().
		the_columns_output_path_is_under_temp("columns.columns.json").and().
		the_columns_file_matches("columns/columns.json")
}
