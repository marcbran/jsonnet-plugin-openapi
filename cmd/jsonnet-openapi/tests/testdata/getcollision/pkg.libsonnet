local p = import 'pkg/main.libsonnet';
p.pkg({
  repo: 'git@github.com:marcbran/jsonnet.git',
  branch: 'openapi/getcollision',
  path: 'openapi-getcollision',
  target: 'getcollision',
}, 'Get collision 1')
