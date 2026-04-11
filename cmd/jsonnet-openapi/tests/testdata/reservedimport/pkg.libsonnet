local p = import 'pkg/main.libsonnet';
p.pkg({
  repo: 'git@github.com:marcbran/jsonnet.git',
  branch: 'openapi/reservedimport',
  path: 'openapi-reservedimport',
  target: 'reservedimport',
}, 'Reserved import path segment 1')
