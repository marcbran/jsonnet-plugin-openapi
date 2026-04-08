local p = import 'pkg/main.libsonnet';
p.pkg({
  repo: 'git@github.com:marcbran/jsonnet.git',
  branch: 'openapi/unsafeident',
  path: 'openapi-unsafeident',
  target: 'unsafeident',
}, 'Unsafe Jsonnet identifiers 1')
