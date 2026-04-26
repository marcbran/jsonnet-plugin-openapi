local p = import 'pkg/main.libsonnet';

p.pkg({
  source: 'https://github.com/marcbran/jsonnet-plugin-openapi',
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'plugin/openapi',
  path: 'plugin/openapi',
  target: 'openapi',
}, |||
  OpenAPI-oriented native functions and code generation support for Jsonnet.
|||)
