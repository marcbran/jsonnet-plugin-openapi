local resourcePathsFor = import 'resource-paths.jsonnet';
local schema = import 'inference-schema.libsonnet';

local pathPartName(part) =
  if std.length(part) >= 2 && part[0] == '{' && part[std.length(part) - 1] == '}'
  then '_' + std.substr(part, 1, std.length(part) - 2)
  else part;
// A bundle name becomes a directory name, so it has to stay inside the
// filesystem's 255-byte limit. Deeply nested paths can exceed it; long names are
// truncated and disambiguated with a digest. Short names are left untouched, so
// existing cached results stay valid.
local boundedName(name) =
  if std.length(name) <= 180 then name
  else std.substr(name, 0, 160) + '--' + std.md5(name);

local bundleName(path) =
  local parts = [pathPartName(part) for part in std.split(path, '/') if part != ''];
  boundedName(if std.length(parts) == 0 then '_root' else std.join('__', parts));

local bundle(spec, sourcePath, detailPaths) =
  local response = std.get(std.get(spec.paths[sourcePath].get, 'responses', {}), '200', null);
  {
    sourcePath: sourcePath,
    responseSchema: schema.resolvedResponseSchema(spec, response),
    detailPaths: [path for path in detailPaths if path != sourcePath],
  };

function(spec)
  local detailPaths = resourcePathsFor(spec);
  {
    [bundleName(sourcePath)]: {
      'input.json': std.manifestJsonEx(bundle(spec, sourcePath, detailPaths), '  '),
    }
    for sourcePath in detailPaths
  }
