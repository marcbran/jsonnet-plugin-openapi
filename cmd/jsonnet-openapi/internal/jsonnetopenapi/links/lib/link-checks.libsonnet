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
    if !std.member(inherited, param)
  ];

local propertiesOf = schema.pathProperties;
local schemaAtInfo = schema.schemaAtPath;

local scalarTypes = ['string', 'integer', 'number'];
local isObjectLike(sch) =
  sch != null && std.type(sch) == 'object' && (
    std.get(sch, 'type', null) == 'object' ||
    std.objectHas(sch, 'properties') ||
    std.any([isObjectLike(branch) for branch in schema.branchesOf(sch)])
  );
local isScalarSchema(sch) =
  sch != null && std.type(sch) == 'object' && std.member(scalarTypes, std.get(sch, 'type', null));
local isValidAt(info) =
  info.schema != null && std.type(info.schema) == 'object' && (
    isObjectLike(info.schema) ||
    (info.crossedArray && isScalarSchema(info.schema))
  );
local isValidRelativePath(spec, atSchema, path) =
  if isScalarSchema(atSchema) then std.length(path) == 0
  else
    local info = schemaAtInfo(spec, atSchema, path);
    isScalarSchema(info.schema) && !info.crossedArray;
local keysValid(spec, atSchema, keys) =
  std.all([isValidRelativePath(spec, atSchema, seg.path) for seg in keys if seg.path != null]);

local hasReachableScalar(sch, depth=8) =
  isScalarSchema(sch) || (
    depth > 0 && isObjectLike(sch) && std.any([
      hasReachableScalar(prop, depth - 1)
      for prop in std.objectValues(propertiesOf(sch))
      if std.get(prop, 'type', null) != 'array'
    ])
  );

local linkProblems(responseSchema, detailPaths, expectedAt, link) =
  local info = schemaAtInfo({}, responseSchema, link.at);
  local missing = missingParams(link.sourcePath, link.targetPath);
  local keyPaths = [seg.path for seg in link.keys if seg.path != null];
  local at = std.manifestJsonMinified(link.at);
  [
    std.format('targetPath %s is not one of detailPaths', [link.targetPath])
    for _ in [null]
    if !std.member(detailPaths, link.targetPath)
  ] + [
    std.format('at %s must be %s, as given in input.json', [at, std.manifestJsonMinified(expectedAt)])
    for _ in [null]
    if expectedAt != null && link.at != expectedAt
  ] + (
    if !isValidAt(info) then [
      std.format('at %s does not resolve to an object in the response schema, or to the scalar items of an array it crosses', [at]),
    ]
    else [
      std.format('keys path %s does not resolve to a scalar relative to at %s without crossing an array', [std.manifestJsonMinified(path), at])
      for path in keyPaths
      if !isValidRelativePath({}, info.schema, path)
    ] + [
      std.format('at %s crosses an array, so keys must end in a path segment that distinguishes the items', [at])
      for _ in [null]
      if info.crossedArray && (std.length(link.keys) == 0 || link.keys[std.length(link.keys) - 1].path == null)
    ] + [
      std.format('at %s has no scalar value reachable without crossing an array, so target params %s cannot be filled; if the values sit inside an array below at, extend at into that array, otherwise drop the link', [at, std.join(', ', missing)])
      for _ in [null]
      if std.length(missing) > 0 && !hasReachableScalar(info.schema)
    ]
  );

local varsProblems(input, output) =
  [
    std.format('param %s is not one of missingParams', [v.param])
    for v in output.vars
    if !std.member(input.missingParams, v.param)
  ] + [
    std.format('vars path %s for param %s does not resolve to a scalar relative to at without crossing an array', [std.manifestJsonMinified(v.path), v.param])
    for v in output.vars
    if std.member(input.missingParams, v.param) && !isValidRelativePath({}, input.itemSchema, v.path)
  ];

local isValidVar(input, v) =
  std.member(input.missingParams, v.param) && isValidRelativePath({}, input.itemSchema, v.path);

{
  pathParams: pathParams,
  inheritedParams: inheritedParams,
  missingParams: missingParams,
  schemaAtInfo: schemaAtInfo,
  isValidAt: isValidAt,
  isValidRelativePath: isValidRelativePath,
  keysValid: keysValid,
  linkProblems: linkProblems,
  varsProblems: varsProblems,
  isValidVar: isValidVar,
}
