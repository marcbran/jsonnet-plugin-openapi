local listPathsFor = import 'list-paths.jsonnet';
local detailPathsFor = import 'detail-paths.jsonnet';
local schema = import 'inference-schema.libsonnet';

local splitPath(path) = [part for part in std.split(path, '/') if part != ''];
local joinPath(parts) = '/' + std.join('/', parts);
local parentPath(path) =
  local parts = splitPath(path);
  if std.length(parts) <= 1 then null else joinPath(std.slice(parts, 0, std.length(parts) - 1, 1));

local uniqueRefs(refs) =
  std.foldl(function(acc, ref) if std.member(acc, ref) then acc else acc + [ref], refs, []);
local contains(values, value) = std.member(values, value);

function(spec)
  local detailPaths = detailPathsFor(spec);
  local listPaths = listPathsFor(spec);
  local mappings = [
    {
      sourcePath: sourcePath,
      targetPath: detailPath,
      array: arrayRef.array,
    }
    for detailPath in detailPaths
    for sourcePath in [parentPath(detailPath)]
    if sourcePath != null && contains(listPaths, sourcePath)
    for sourceResponse in [std.get(std.get(spec.paths[sourcePath].get, 'responses', {}), '200', null)]
    for sourceSchema in [schema.resolvedResponseSchema(spec, sourceResponse)]
    for sourceArrayRefs in [schema.arrayRefs(sourceSchema)]
    if std.length(sourceArrayRefs) == 1
    for arrayRef in sourceArrayRefs
  ];
  local mappedListPaths = uniqueRefs([mapping.sourcePath for mapping in mappings]);
  {
    mappings: mappings,
    unmappedListPaths: [
      path
      for path in listPaths
      if !contains(mappedListPaths, path)
    ],
    detailPaths: detailPaths,
  }
