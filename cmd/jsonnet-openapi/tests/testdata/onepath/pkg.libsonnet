local p = import 'pkg/main.libsonnet';
p.pkg({
  repo: 'git@github.com:marcbran/jsonnet.git',
  branch: 'openapi/onepath',
  path: 'openapi-onepath',
  target: 'onepath',
}, 'One 1')
