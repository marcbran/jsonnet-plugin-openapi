//go:build e2e

package tests

import "testing"

func TestListColumnsInferWithCachedInference(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_list_columns_spec_file("listcolumns.yaml").and().
		a_list_columns_output_under_temp("listcolumns.columns.json").and().
		a_list_columns_workdir_under_temp("listcolumns-work").and().
		a_cached_user_columns_inference()

	when.
		the_list_columns_command_is_run()

	then.
		the_list_columns_has_no_error().and().
		the_columns_output_path_is_under_temp("listcolumns.columns.json").and().
		the_columns_file_matches("listcolumns/columns.json")
}
