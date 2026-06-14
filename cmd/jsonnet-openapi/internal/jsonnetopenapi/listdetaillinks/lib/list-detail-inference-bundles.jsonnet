local nestedMappingsFor = import 'list-detail-nested.jsonnet';
local schema = import 'inference-schema.libsonnet';

local pathPartName(part) =
  if std.length(part) >= 2 && part[0] == '{' && part[std.length(part) - 1] == '}'
  then '_' + std.substr(part, 1, std.length(part) - 2)
  else part;
local bundleName(path) =
  local parts = [pathPartName(part) for part in std.split(path, '/') if part != ''];
  if std.length(parts) == 0 then '_root' else std.join('__', parts);

local listPathBundle(spec, result, sourcePath) =
  local response = std.get(std.get(spec.paths[sourcePath].get, 'responses', {}), '200', null);
  local responseSchema = schema.resolvedResponseSchema(spec, response);
  {
    sourcePath: sourcePath,
    arrays: schema.responseArrays(spec, responseSchema),
    detailPaths: result.detailPaths,
  };

function(spec)
  local result = nestedMappingsFor(spec);
  {
    [bundleName(sourcePath)]: {
      'input.json': std.manifestJsonEx(
        listPathBundle(spec, result, sourcePath),
        '  ',
      ),
    }
    for sourcePath in result.unmappedListPaths
  }
