//go:build e2e

package tests

import "testing"

func TestLinksInferWithCachedInference(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_links_spec_file("links.yaml").and().
		a_links_output_under_temp("links.links.json").and().
		a_links_workdir_under_temp("links-work").and().
		a_cached_account_link_from_resource().and().
		a_cached_user_link_from_resource().and().
		a_cached_users_link_from_collection().and().
		a_cached_reports_link_from_collection().and().
		a_cached_member_link_vars().and().
		a_cached_users_link_vars()

	when.
		the_links_command_is_run()

	then.
		the_links_has_no_error().and().
		the_links_output_path_is_under_temp("links.links.json").and().
		the_links_file_matches("links/links.json")
}
