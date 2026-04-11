local p = import 'pkg/main.libsonnet';
p.pkg({
  repo: 'git@github.com:marcbran/jsonnet.git',
  branch: 'openapi/multiparam',
  path: 'openapi-multiparam',
  target: 'multiparam',
}, 'Multiparam 1')
