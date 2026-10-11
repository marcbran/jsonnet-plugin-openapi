local checks = import 'link-checks.libsonnet';
local schema = import 'inference-schema.libsonnet';

local splitPath(path) = [part for part in std.split(path, '/') if part != ''];
local pathParam(part) =
  if std.length(part) >= 2 && part[0] == '{' && part[std.length(part) - 1] == '}'
  then std.substr(part, 1, std.length(part) - 2)
  else null;
local schemaAtInfo = checks.schemaAtInfo;
local isValidAt = checks.isValidAt;
local isValidRelativePath = checks.isValidRelativePath;

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
  local inherited = checks.inheritedParams(sourcePath, targetPath);
  [
    local param = pathParam(seg);
    if param == null then { const: seg }
    else if std.member(inherited, param) then { origin: param }
    else { param: param, path: varPath(vars, param) }
    for seg in splitPath(targetPath)
  ];

local isResolved(segments) =
  std.all([
    !(std.objectHas(seg, 'path') && seg.path == null)
    for seg in segments
  ]);

local normalizeLabel(label) = std.strReplace(std.strReplace(label, '-', ' '), '_', ' ');
local cleanKeySegment(seg) =
  if seg.const != null then { const: normalizeLabel(seg.const) } else { path: seg.path };
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
