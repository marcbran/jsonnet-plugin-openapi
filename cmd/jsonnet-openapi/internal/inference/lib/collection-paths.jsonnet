local endsWithString(value, suffix) =
  std.length(value) >= std.length(suffix) &&
  std.substr(value, std.length(value) - std.length(suffix), std.length(suffix)) == suffix;
local isWatchPath(path) =
  std.findSubstr('/watch/', path) != [];
local isKubernetesDiscoveryOperation(operation) =
  local operationId = std.get(operation, 'operationId', '');
  std.startsWith(operationId, 'get') && endsWithString(operationId, 'APIResources');
local schema = import 'inference-schema.libsonnet';

// Collection endpoints, identified purely from the response shape. Note there
// is deliberately no test on the path: a collection can sit under a path
// parameter (/teams/{id}/members, /standards/scores/{resource_type}), and a
// parameterless path can return a single resource (/user).
function(spec)
  [
    path
    for path in std.objectFields(spec.paths)
    if std.objectHas(spec.paths[path], 'get')
    for operation in [spec.paths[path].get]
    if !isWatchPath(path)
    if !isKubernetesDiscoveryOperation(operation)
    for response in [std.get(std.get(operation, 'responses', {}), '200', null)]
    if response != null
    for responseSchema in [schema.resolvedResponseSchema(spec, response)]
    if schema.isCollectionResponseSchema(responseSchema)
  ]
