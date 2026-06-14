local listPathsFor = import 'list-paths.jsonnet';
local schema = import 'inference-schema.libsonnet';

local pathPartName(part) =
  if std.length(part) >= 2 && part[0] == '{' && part[std.length(part) - 1] == '}'
  then '_' + std.substr(part, 1, std.length(part) - 2)
  else part;
local bundleName(path) =
  local parts = [pathPartName(part) for part in std.split(path, '/') if part != ''];
  if std.length(parts) == 0 then '_root' else std.join('__', parts);
local bundleNameFor(path, array) =
  local name = bundleName(path);
  if std.length(array) == 0 then name else name + '___' + std.join('__', array);

local listPathBundle(spec, sourcePath, array) = {
  sourcePath: sourcePath,
  operationId: std.get(spec.paths[sourcePath].get, 'operationId', null),
  array: array.array,
  itemSchema: array.itemSchema,
};

function(spec)
  {
    [bundleNameFor(sourcePath, array.array)]: {
      'input.json': std.manifestJsonEx(
        listPathBundle(spec, sourcePath, array),
        '  ',
      ),
    }
    for sourcePath in listPathsFor(spec)
    for response in [std.get(std.get(spec.paths[sourcePath].get, 'responses', {}), '200', null)]
    for responseSchema in [schema.resolvedResponseSchema(spec, response)]
    for array in schema.responseArrays(spec, responseSchema)
  }
