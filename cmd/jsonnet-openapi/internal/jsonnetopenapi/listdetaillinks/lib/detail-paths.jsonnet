local endsWithPathVar(path) =
  std.length(path) >= 2 &&
  path[std.length(path) - 1] == '}' &&
  std.findSubstr('{', path) != [];
local isWatchPath(path) =
  std.findSubstr('/watch/', path) != [];
local schema = import 'inference-schema.libsonnet';

local isSingleRefSchema(schema) =
  std.type(schema) == 'object' &&
  std.objectFields(schema) == ['$ref'];
local isOneOfRefSchema(schema) =
  std.type(schema) == 'object' &&
  std.objectHas(schema, 'oneOf') &&
  std.length(schema.oneOf) > 0 &&
  std.all([isSingleRefSchema(item) for item in schema.oneOf]);
local isDetailResponseSchema(schema) =
  isSingleRefSchema(schema) || isOneOfRefSchema(schema);

function(spec)
  [
    path
    for path in std.objectFields(spec.paths)
    if std.objectHas(spec.paths[path], 'get')
    if !isWatchPath(path)
    if endsWithPathVar(path)
    for response in [std.get(std.get(spec.paths[path].get, 'responses', {}), '200', null)]
    if response != null
    for responseSchema in [schema.responseSchema(spec, response)]
    if isDetailResponseSchema(responseSchema)
  ]
