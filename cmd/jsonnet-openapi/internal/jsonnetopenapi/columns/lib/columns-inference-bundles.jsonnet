local collectionPathsFor = import 'collection-paths.jsonnet';
local schema = import 'inference-schema.libsonnet';

local pathPartName(part) =
  if std.length(part) >= 2 && part[0] == '{' && part[std.length(part) - 1] == '}'
  then '_' + std.substr(part, 1, std.length(part) - 2)
  else part;
local bundleName(path) =
  local parts = [pathPartName(part) for part in std.split(path, '/') if part != ''];
  if std.length(parts) == 0 then '_root' else std.join('__', parts);
// A bundle name becomes a directory name, so it has to stay inside the
// filesystem's 255-byte limit. Deeply nested paths can exceed it; long names are
// truncated and disambiguated with a digest. Short names are left untouched, so
// existing cached results stay valid.
local boundedName(name) =
  if std.length(name) <= 180 then name
  else std.substr(name, 0, 160) + '--' + std.md5(name);

local bundleNameFor(path, array) =
  local name = bundleName(path);
  boundedName(if std.length(array) == 0 then name else name + '___' + std.join('__', array));

local listPathBundle(spec, sourcePath, itemsPath, itemSchema) = {
  sourcePath: sourcePath,
  operationId: std.get(spec.paths[sourcePath].get, 'operationId', null),
  array: itemsPath,
  itemSchema: itemSchema,
};

// A collection has exactly one item array, so each source path yields one
// bundle.
function(spec)
  {
    [bundleNameFor(sourcePath, itemsPath)]: {
      'input.json': std.manifestJsonEx(
        listPathBundle(spec, sourcePath, itemsPath, itemSchema),
        '  ',
      ),
    }
    for sourcePath in collectionPathsFor(spec)
    for response in [std.get(std.get(spec.paths[sourcePath].get, 'responses', {}), '200', null)]
    for responseSchema in [schema.resolvedResponseSchema(spec, response)]
    for itemsPath in [schema.collectionItemsPath(responseSchema)]
    if itemsPath != null
    for itemSchema in [schema.arrayItemSchema(spec, responseSchema, itemsPath)]
  }
