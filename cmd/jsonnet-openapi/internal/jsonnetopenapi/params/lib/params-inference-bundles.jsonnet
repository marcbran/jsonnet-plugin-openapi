local collectionPathsFor = import 'collection-paths.jsonnet';
local schema = import 'inference-schema.libsonnet';

local pathPartName(part) =
  if std.length(part) >= 2 && part[0] == '{' && part[std.length(part) - 1] == '}'
  then '_' + std.substr(part, 1, std.length(part) - 2)
  else part;
local bundleName(path) =
  local parts = [pathPartName(part) for part in std.split(path, '/') if part != ''];
  if std.length(parts) == 0 then '_root' else std.join('__', parts);
local boundedName(name) =
  if std.length(name) <= 180 then name
  else std.substr(name, 0, 160) + '--' + std.md5(name);

local parameterRefName(ref) = std.strReplace(ref, '#/components/parameters/', '');
local resolveParameter(spec, parameter) =
  if std.objectHas(parameter, '$ref')
  then std.get(std.get(std.get(spec, 'components', {}), 'parameters', {}), parameterRefName(parameter['$ref']), {})
  else parameter;

local queryParameters(spec, path) =
  local item = spec.paths[path];
  local resolved = [
    resolveParameter(spec, parameter)
    for parameter in std.get(item, 'parameters', []) + std.get(item.get, 'parameters', [])
  ];
  local byName = std.foldl(
    function(acc, parameter) acc { [parameter.name]: parameter },
    [parameter for parameter in resolved if std.get(parameter, 'in', null) == 'query'],
    {},
  );
  [byName[name] for name in std.objectFields(byName)];

local compactSchema(spec, parameter) =
  local resolved = schema.resolveSchema(spec, std.get(parameter, 'schema', {}));
  local value = if resolved == null then {} else resolved;
  local items = if std.get(value, 'type', null) == 'array' then std.get(value, 'items', null) else null;
  {
    [field]: value[field]
    for field in ['type', 'format', 'enum', 'default']
    if std.objectHas(value, field)
  } + (
    if items == null then {}
    else { items: { [field]: items[field] for field in ['type', 'format', 'enum'] if std.objectHas(items, field) } }
  );

local paramEntry(spec, parameter) = {
  name: parameter.name,
  required: std.get(parameter, 'required', false),
  schema: compactSchema(spec, parameter),
} + (
  if std.objectHas(parameter, 'description') then { description: parameter.description } else {}
);

local paramsBundle(spec, sourcePath, params) =
  local operation = spec.paths[sourcePath].get;
  {
    sourcePath: sourcePath,
    operationId: std.get(operation, 'operationId', null),
    params: [paramEntry(spec, parameter) for parameter in params],
  } + (
    if std.objectHas(operation, 'summary') then { summary: operation.summary } else {}
  );

function(spec)
  {
    [boundedName(bundleName(sourcePath))]: {
      'input.json': std.manifestJsonEx(paramsBundle(spec, sourcePath, params), '  '),
    }
    for sourcePath in collectionPathsFor(spec)
    for params in [queryParameters(spec, sourcePath)]
    if std.length(params) > 0
  }
