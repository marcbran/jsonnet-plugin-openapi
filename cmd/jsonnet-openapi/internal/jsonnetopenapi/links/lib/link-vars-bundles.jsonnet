local checks = import 'link-checks.libsonnet';
local resourcePathsFor = import 'resource-paths.jsonnet';
local schema = import 'inference-schema.libsonnet';

local splitPath(path) = [part for part in std.split(path, '/') if part != ''];
local pathParams = checks.pathParams;
local inheritedParams = checks.inheritedParams;
local missingParams = checks.missingParams;

local pathPartName(part) =
  if std.length(part) >= 2 && part[0] == '{' && part[std.length(part) - 1] == '}'
  then '_' + std.substr(part, 1, std.length(part) - 2)
  else part;
local bundlePathName(path) =
  local parts = [pathPartName(part) for part in splitPath(path)];
  if std.length(parts) == 0 then '_root' else std.join('__', parts);

local sanitizeSlug(s) =
  std.join('', [
    if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-'
    then c
    else '_'
    for c in std.stringChars(s)
  ]);
local keySegmentSlug(seg) =
  if seg.const != null then sanitizeSlug(seg.const)
  else 'p_' + std.join('_', [sanitizeSlug(p) for p in seg.path]);
local keysSlug(keys) = std.join('__', [keySegmentSlug(seg) for seg in keys]);

// A bundle name becomes a directory name, so it has to stay inside the
// filesystem's 255-byte limit. Deeply nested paths blow past that once source
// and target are concatenated, so long names are truncated and disambiguated
// with a digest of the full name. Short names are left untouched, which keeps
// existing cached results valid.
local boundedName(name) =
  if std.length(name) <= 180 then name
  else std.substr(name, 0, 160) + '--' + std.md5(name);

local bundleName(link) =
  boundedName(
    bundlePathName(link.sourcePath) + '--' + bundlePathName(link.targetPath) +
    (if std.length(link.at) == 0 then '' else '--at_' + std.join('_', link.at)) +
    '--k_' + keysSlug(link.keys)
  );

local schemaAtInfo = checks.schemaAtInfo;

local bundle(spec, link, itemSchema) =
  {
    sourcePath: link.sourcePath,
    targetPath: link.targetPath,
    at: link.at,
    keys: link.keys,
    sourceParams: pathParams(link.sourcePath),
    inheritedParams: inheritedParams(link.sourcePath, link.targetPath),
    targetParams: pathParams(link.targetPath),
    missingParams: missingParams(link.sourcePath, link.targetPath),
    itemSchema: itemSchema,
  };

function(spec, inferred=(import 'links-from-resources/results/all.jsonnet') + (import 'links-from-collections/results/all.jsonnet'))
  local links = std.flattenArrays([result.links for result in inferred]);
  local detailPaths = resourcePathsFor(spec);
  {
    [bundleName(link)]: {
      'input.json': std.manifestJsonEx(bundle(spec, link, info.schema), '  '),
    }
    for link in links
    for response in [std.get(std.get(spec.paths[link.sourcePath].get, 'responses', {}), '200', null)]
    for responseSchema in [schema.resolvedResponseSchema(spec, response)]
    for info in [schemaAtInfo(spec, responseSchema, link.at)]
    if std.length(checks.linkProblems(responseSchema, detailPaths, null, link)) == 0 &&
       std.length(missingParams(link.sourcePath, link.targetPath)) > 0
  }
