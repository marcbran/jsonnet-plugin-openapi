//go:build e2e

package tests

import "testing"

func TestAPISpec(t *testing.T) {
	cases := []string{
		"minimal",
		"getcollision",
		"onepath",
		"paramonly",
		"multiparam",
		"unsafeident",
		"reservedimport",
	}
	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			given, when, then := specScenario(t)

			given.
				a_fixture_spec(tc)

			when.
				the_jsonnet_api_spec_is_evaluated()

			then.
				the_eval_has_no_error().and().
				the_result_matches_expected_file(tc, "api.json")
		})
	}
}

func TestNestedSpec(t *testing.T) {
	cases := []string{
		"minimal",
		"getcollision",
		"onepath",
		"paramonly",
		"multiparam",
		"unsafeident",
		"reservedimport",
	}
	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			given, when, then := specScenario(t)

			given.
				a_fixture_spec(tc)

			when.
				the_jsonnet_nested_spec_is_evaluated()

			then.
				the_eval_has_no_error().and().
				the_result_matches_expected_file(tc, "nested.json")
		})
	}
}
