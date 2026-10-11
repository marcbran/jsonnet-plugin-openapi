local checks = import 'link-checks.libsonnet';

function(input, output)
  local expectedAt = std.get(input, 'at', null);
  local checked = [
    { link: link, problems: checks.linkProblems(input.responseSchema, input.detailPaths, expectedAt, link) }
    for link in output.links
  ];
  {
    valid: output { links: [entry.link for entry in checked if std.length(entry.problems) == 0] },
    problems: std.flattenArrays([entry.problems for entry in checked]),
  }
