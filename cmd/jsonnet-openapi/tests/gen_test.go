//go:build e2e

package tests

import "testing"

func TestGenerate(t *testing.T) {
	cases := []struct {
		name    string
		spec    string
		service string
	}{
		{name: "minimal", spec: "minimal", service: "minimal"},
		{name: "getcollision", spec: "getcollision", service: "getcollision"},
		{name: "onepath", spec: "onepath", service: "onepath"},
		{name: "paramonly", spec: "paramonly", service: "paramonly"},
		{name: "multiparam", spec: "multiparam", service: "multiparam"},
		{name: "unsafeident", spec: "unsafeident", service: "unsafeident"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			given, when, then := scenario(t)

			given.
				a_spec(tc.spec).and().
				a_service(tc.service)

			when.
				the_gen_command_is_run()

			then.
				the_gen_has_no_error().and().
				the_generated_files_match(tc.spec).and().
				the_generated_main_libsonnet_parses_as_jsonnet()
		})
	}
}
