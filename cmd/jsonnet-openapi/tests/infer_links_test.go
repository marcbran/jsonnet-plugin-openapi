//go:build e2e

package tests

import "testing"

func TestInferLinksWithCachedVarInference(t *testing.T) {
	given, when, then := scenario(t)

	given.
		an_infer_links_spec("listdetail").and().
		an_infer_links_output_under_temp("listdetail.links.json").and().
		an_infer_links_workdir_under_temp("listdetail-work").and().
		a_cached_user_detail_var_inference()

	when.
		the_infer_links_command_is_run()

	then.
		the_infer_links_has_no_error().and().
		the_links_output_path_is_under_temp("listdetail.links.json").and().
		the_links_file_contains("[\n   {\n      \"array\": [ ],\n      \"sourcePath\": \"/users\",\n      \"targetPath\": \"/users/{userId}\",\n      \"vars\": {\n         \"userId\": [\n            \"id\"\n         ]\n      }\n   }\n]\n")
}
