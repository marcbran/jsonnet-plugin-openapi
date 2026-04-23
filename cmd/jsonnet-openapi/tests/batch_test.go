//go:build e2e

package tests

import "testing"

func TestBatch_codegen_jobs_run_in_order(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_batch_job_from_testdata("minimal", "out-a", "a").and().
		a_batch_job_from_testdata("minimal", "out-b", "b")

	when.
		the_batch_command_is_run()

	then.
		the_batch_has_no_error().and().
		the_batch_job_count_is(2).and().
		the_batch_output_out_dirs_under_temp_are("out-a", "out-b").and().
		the_generated_main_libsonnet_exists_for_batch_job(0).and().
		the_generated_main_libsonnet_exists_for_batch_job(1)
}

func TestBatch_codegen_stops_when_a_job_fails(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_batch_job_from_testdata("minimal", "ok1", "ok1").and().
		a_batch_job_with_missing_spec_file("fail-out", "fail").and().
		a_batch_job_from_testdata("minimal", "skipped", "skipped")

	when.
		the_batch_command_is_run()

	then.
		the_batch_has_an_error().and().
		the_batch_outputs_are_nil().and().
		the_generated_main_libsonnet_exists_under_temp("ok1").and().
		the_generated_main_libsonnet_does_not_exist_under_temp("skipped")
}
