//go:build e2e

package tests

import "testing"

func TestParamsInferWithCachedInference(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_params_spec_file("params.yaml").and().
		a_params_output_under_temp("params.params.json").and().
		a_params_workdir_under_temp("params-work").and().
		a_cached_user_params_inference()

	when.
		the_params_command_is_run()

	then.
		the_params_has_no_error().and().
		the_params_output_path_is_under_temp("params.params.json").and().
		the_params_file_matches("params/params.json")
}
