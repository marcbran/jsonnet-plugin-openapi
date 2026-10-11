local checks = import 'link-checks.libsonnet';

function(input, output)
  {
    valid: output { vars: [v for v in output.vars if checks.isValidVar(input, v)] },
    problems: checks.varsProblems(input, output),
  }
