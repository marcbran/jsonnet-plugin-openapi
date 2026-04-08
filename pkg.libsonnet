local p = import 'pkg/main.libsonnet';

p.pkg({
  source: 'https://github.com/marcbran/jsonnet-plugin-openapi',
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'plugin-openapi',
  path: 'plugin/openapi',
  target: 'openapi',
}, |||
  Read-only HTTP GET requests against a REST API. Base URL and default headers are configured when the plugin is started or embedded in Go.
|||, {
  request: p.desc(|||
    Sends a GET request. `input` is an object with `method` (`GET` only), `path`, optional `headers`, and optional `params` (query string).

    On success returns parsed JSON. On failure returns a `Status` object (`kind: "Status"`).
  |||),
})
