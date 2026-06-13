//go:build e2e

package tests

import "testing"

func TestListDetailLinksInferWithCachedVarInference(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_list_detail_links_spec_file("listdetaillinks.yaml").and().
		a_list_detail_links_output_under_temp("listdetaillinks.links.json").and().
		a_list_detail_links_workdir_under_temp("listdetaillinks-work").and().
		a_cached_user_detail_var_inference()

	when.
		the_list_detail_links_command_is_run()

	then.
		the_list_detail_links_has_no_error().and().
		the_links_output_path_is_under_temp("listdetaillinks.links.json").and().
		the_links_file_matches("listdetaillinks/links.json")
}
