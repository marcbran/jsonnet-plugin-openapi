local endsWithString(value, suffix) =
  std.length(value) >= std.length(suffix) &&
  std.substr(value, std.length(value) - std.length(suffix), std.length(suffix)) == suffix;
local isWatchPath(path) =
  std.findSubstr('/watch/', path) != [];
local isKubernetesDiscoveryOperation(operation) =
  local operationId = std.get(operation, 'operationId', '');
  std.startsWith(operationId, 'get') && endsWithString(operationId, 'APIResources');
local schema = import 'inference-schema.libsonnet';

// Every GET endpoint that is not a collection is a resource -- there is no
// third category and nothing is left unclassified. In particular a resource
// need not sit under a path parameter (/user, /account/details) and its payload
// need not be $ref-based; specs that inline their schemas are common.
//
// Endpoints whose 200 has no usable JSON schema are excluded: there is nothing
// to infer links from. Consumers still treat them as resources, since anything
// absent from the collection set is one.
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
    if responseSchema != null
    if std.type(responseSchema) == 'object'
    if !schema.isCollectionResponseSchema(responseSchema)
    if !schema.isArraySchema(responseSchema)
  ]
