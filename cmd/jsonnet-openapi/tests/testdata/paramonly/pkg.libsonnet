local p = import 'pkg/main.libsonnet';
p.pkg({
  repo: 'git@github.com:marcbran/jsonnet.git',
  branch: 'openapi/paramonly',
  path: 'openapi-paramonly',
  target: 'paramonly',
}, 'P 1')
