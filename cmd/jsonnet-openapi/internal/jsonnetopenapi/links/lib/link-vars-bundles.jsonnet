local schema = import 'inference-schema.libsonnet';

local splitPath(path) = [part for part in std.split(path, '/') if part != ''];
local pathParam(part) =
  if std.length(part) >= 2 && part[0] == '{' && part[std.length(part) - 1] == '}'
  then std.substr(part, 1, std.length(part) - 2)
  else null;
local pathParams(path) = [
  param
  for part in splitPath(path)
  for param in [pathParam(part)]
  if param != null
];
local contains(values, value) = std.member(values, value);
local commonPrefixLength(left, right) =
  if std.length(left) == 0 || std.length(right) == 0 || left[0] != right[0] then 0
  else 1 + commonPrefixLength(
    std.slice(left, 1, std.length(left), 1),
    std.slice(right, 1, std.length(right), 1),
  );
local inheritedParams(sourcePath, targetPath) =
  local sourceParts = splitPath(sourcePath);
  local targetParts = splitPath(targetPath);
  local commonLength = commonPrefixLength(sourceParts, targetParts);
  local targetIsSourceParent = commonLength == std.length(targetParts) && std.length(sourceParts) > std.length(targetParts);
  if targetIsSourceParent then [
    param
    for part in targetParts
    for param in [pathParam(part)]
    if param != null
  ]
  else [
    param
    for i in std.range(0, commonLength - 1)
    for param in [pathParam(targetParts[i])]
    if param != null
  ];
local missingParams(sourcePath, targetPath) =
  local inherited = inheritedParams(sourcePath, targetPath);
  [
    param
    for param in pathParams(targetPath)
    if !contains(inherited, param)
  ];

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

local propertyOf(sch, name) =
  if sch == null then null
  else if std.objectHas(sch, 'properties') && std.objectHas(sch.properties, name) then sch.properties[name]
  else if std.objectHas(sch, 'allOf') then
    std.foldl(function(acc, branch) if acc != null then acc else propertyOf(branch, name), sch.allOf, null)
  else null;

local schemaAtInfo(spec, sch, at, crossedArray=false) =
  if sch == null then { schema: sch, crossedArray: crossedArray }
  else if std.length(at) == 0 then
    if std.get(sch, 'type', null) == 'array'
    then { schema: schema.resolveSchema(spec, std.get(sch, 'items', null)), crossedArray: true }
    else { schema: sch, crossedArray: crossedArray }
  else
    local propSchema = propertyOf(sch, at[0]);
    local isArray = propSchema != null && std.get(propSchema, 'type', null) == 'array';
    local next =
      if propSchema == null then null
      else if isArray then schema.resolveSchema(spec, std.get(propSchema, 'items', null))
      else schema.resolveSchema(spec, propSchema);
    schemaAtInfo(spec, next, std.slice(at, 1, std.length(at), 1), crossedArray || isArray);
local scalarTypes = ['string', 'integer', 'number'];
local isObjectLike(sch) =
  sch != null && std.type(sch) == 'object' && (
    std.get(sch, 'type', null) == 'object' ||
    std.objectHas(sch, 'properties') ||
    (std.objectHas(sch, 'allOf') && std.any([isObjectLike(branch) for branch in sch.allOf]))
  );
local isValidAt(info) =
  info.schema != null && std.type(info.schema) == 'object' && (
    isObjectLike(info.schema) ||
    (info.crossedArray && std.member(scalarTypes, std.get(info.schema, 'type', null)))
  );
local isScalarSchema(sch) =
  sch != null && std.type(sch) == 'object' && std.member(scalarTypes, std.get(sch, 'type', null));
local isValidRelativePath(spec, atSchema, path) =
  if isScalarSchema(atSchema) then std.length(path) == 0
  else
    local info = schemaAtInfo(spec, atSchema, path);
    isScalarSchema(info.schema) && !info.crossedArray;
local keysValid(spec, atSchema, keys) =
  std.all([isValidRelativePath(spec, atSchema, seg.path) for seg in keys if seg.path != null]);

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
  {
    [bundleName(link)]: {
      'input.json': std.manifestJsonEx(bundle(spec, link, info.schema), '  '),
    }
    for link in links
    for response in [std.get(std.get(spec.paths[link.sourcePath].get, 'responses', {}), '200', null)]
    for responseSchema in [schema.resolvedResponseSchema(spec, response)]
    for info in [schemaAtInfo(spec, responseSchema, link.at)]
    if isValidAt(info) &&
       keysValid(spec, info.schema, link.keys) &&
       std.length(missingParams(link.sourcePath, link.targetPath)) > 0
  }
