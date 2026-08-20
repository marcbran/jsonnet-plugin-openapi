local schema = import 'inference-schema.libsonnet';

local splitPath(path) = [part for part in std.split(path, '/') if part != ''];
local pathParam(part) =
  if std.length(part) >= 2 && part[0] == '{' && part[std.length(part) - 1] == '}'
  then std.substr(part, 1, std.length(part) - 2)
  else null;
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

local matchingVars(varsInferred, link) = [
  item
  for item in varsInferred
  if item.sourcePath == link.sourcePath &&
     item.targetPath == link.targetPath &&
     item.at == link.at &&
     item.keys == link.keys
];

local varPath(vars, param) =
  local matches = [item.path for item in vars if item.param == param];
  if std.length(matches) > 0 then matches[0] else null;

local valueSegments(sourcePath, targetPath, vars) =
  local inherited = inheritedParams(sourcePath, targetPath);
  [
    local param = pathParam(seg);
    if param == null then { const: seg }
    else if contains(inherited, param) then { origin: param }
    else { param: param, path: varPath(vars, param) }
    for seg in splitPath(targetPath)
  ];

local isResolved(segments) =
  std.all([
    !(std.objectHas(seg, 'path') && seg.path == null)
    for seg in segments
  ]);

local cleanKeySegment(seg) =
  if seg.const != null then { const: seg.const } else { path: seg.path };
local cleanKeys(keys) = [cleanKeySegment(seg) for seg in keys];

function(spec, inferred, varsInferred)
  local links = std.flattenArrays([result.links for result in inferred]);
  [
    {
      sourcePath: link.sourcePath,
      at: link.at,
      keys: cleanKeys(link.keys),
      value: value,
    }
    for link in links
    for response in [std.get(std.get(spec.paths[link.sourcePath].get, 'responses', {}), '200', null)]
    for responseSchema in [schema.resolvedResponseSchema(spec, response)]
    for atInfo in [schemaAtInfo(spec, responseSchema, link.at)]
    if isValidAt(atInfo)
    if std.all([isValidRelativePath(spec, atInfo.schema, seg.path) for seg in link.keys if seg.path != null])
    for matches in [matchingVars(varsInferred, link)]
    for vars in [if std.length(matches) == 0 then [] else matches[0].vars]
    for value in [valueSegments(link.sourcePath, link.targetPath, vars)]
    if isResolved(value)
    if std.all([isValidRelativePath(spec, atInfo.schema, seg.path) for seg in value if std.objectHas(seg, 'path')])
  ]
