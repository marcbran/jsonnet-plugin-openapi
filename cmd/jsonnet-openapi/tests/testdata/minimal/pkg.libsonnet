local p = import 'pkg/main.libsonnet';
p.pkg({
  repo: 'git@github.com:marcbran/jsonnet.git',
  branch: 'openapi/minimal',
  path: 'openapi-minimal',
  target: 'minimal',
}, 'Minimal 1')
