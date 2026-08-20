local responseRefName(ref) = std.strReplace(ref, '#/components/responses/', '');

local unescapeRefSegment(segment) =
  std.strReplace(std.strReplace(segment, '~1', '/'), '~0', '~');

local resolveSchemaRef(spec, ref) =
  local path = std.strReplace(ref, '#/components/schemas/', '');
  local segments = [unescapeRefSegment(s) for s in std.split(path, '/')];
  std.foldl(
    function(node, segment)
      if std.type(node) == 'array' then node[std.parseInt(segment)]
      else node[segment],
    segments,
    spec.components.schemas
  );

local responseSchema(spec, response) =
  local resolved =
    if std.type(response) == 'object' &&
       std.objectHas(response, '$ref') &&
       std.startsWith(response['$ref'], '#/components/responses/')
    then spec.components.responses[responseRefName(response['$ref'])]
    else response;
  local content = std.get(resolved, 'content', {});
  local contentTypes = std.objectFields(content);
  local child =
    if std.objectHas(content, 'application/json') then content['application/json']
    else if std.length(contentTypes) > 0 then content[contentTypes[0]]
    else {};
  std.get(child, 'schema', null);

local isSingleRefSchema(schema) =
  std.type(schema) == 'object' &&
  std.objectFields(schema) == ['$ref'];

local isAllOfRefSchema(schema) =
  std.type(schema) == 'object' &&
  std.objectHas(schema, 'allOf') &&
  std.length(schema.allOf) == 1 &&
  isSingleRefSchema(schema.allOf[0]);

local refString(schema) =
  if isSingleRefSchema(schema) && std.startsWith(schema['$ref'], '#/components/schemas/')
  then schema['$ref']
  else if isAllOfRefSchema(schema) && std.startsWith(schema.allOf[0]['$ref'], '#/components/schemas/')
  then schema.allOf[0]['$ref']
  else null;

local normalizeSchema(spec, schema, seen=[]) =
  local ref = refString(schema);
  if ref != null then
    if std.member(seen, ref) then {
      'x-jsonnet-openapi-ref': ref,
      'x-jsonnet-openapi-recursiveRef': true,
    } else
      normalizeSchema(spec, resolveSchemaRef(spec, ref), seen + [ref]) {
        'x-jsonnet-openapi-ref': ref,
      }
  else if std.type(schema) == 'object' then {
    [field]:
      if field == 'items' then normalizeSchema(spec, schema[field], seen)
      else if field == 'properties' then {
        [property]: normalizeSchema(spec, schema[field][property], seen)
        for property in std.objectFields(schema[field])
      }
      else if field == 'oneOf' || field == 'anyOf' || field == 'allOf' then [
        normalizeSchema(spec, item, seen)
        for item in schema[field]
      ]
      else schema[field]
    for field in std.objectFields(schema)
  }
  else schema;

local resolveSchema(spec, schema) =
  local ref = refString(schema);
  if ref != null then normalizeSchema(spec, resolveSchemaRef(spec, ref), [ref]) {
    'x-jsonnet-openapi-ref': ref,
  } else normalizeSchema(spec, schema);

local resolvedResponseSchema(spec, response) =
  resolveSchema(spec, responseSchema(spec, response));

local isArraySchema(schema) =
  std.type(schema) == 'object' &&
  std.get(schema, 'type', null) == 'array';

local effectiveProperties(schema) =
  if std.type(schema) != 'object' then {}
  else
    local fromAllOf =
      if std.objectHas(schema, 'allOf') && std.type(schema.allOf) == 'array' then
        std.foldl(
          function(acc, branch) acc + effectiveProperties(branch),
          schema.allOf,
          {}
        )
      else {};
    fromAllOf + std.get(schema, 'properties', {});

local identityFields = [
  'html_url',
  'id',
  'node_id',
  'self',
  'url',
];

local isObjectLikeSchema(schema) =
  std.type(schema) == 'object' && (
    std.get(schema, 'type', null) == 'object' ||
    std.objectHas(schema, 'properties') ||
    (std.objectHas(schema, 'allOf') &&
     std.any([isObjectLikeSchema(branch) for branch in schema.allOf]))
  );

local collectionItemsPath(schema) =
  if schema == null || std.type(schema) != 'object' then null
  else if isArraySchema(schema) then []
  else
    local properties = effectiveProperties(schema);
    local names = std.objectFields(properties);
    local arrayNames = [name for name in names if isArraySchema(properties[name])];
    local objectArrayNames = [
      name
      for name in arrayNames
      if isObjectLikeSchema(std.get(properties[name], 'items', null))
    ];
    local hasIdentity = std.any([std.member(identityFields, name) for name in names]);
    if hasIdentity then null
    else if std.length(objectArrayNames) == 1 then [objectArrayNames[0]]
    else if std.length(objectArrayNames) == 0 && std.length(arrayNames) == 1 then [arrayNames[0]]
    else null;

local isCollectionResponseSchema(schema) = collectionItemsPath(schema) != null;

local arrayItemSchema(spec, schema, array) =
  if std.length(array) == 0 then
    resolveSchema(spec, std.get(schema, 'items', null))
  else if std.length(array) == 1 then
    local properties = effectiveProperties(schema);
    if std.objectHas(properties, array[0])
    then resolveSchema(spec, std.get(properties[array[0]], 'items', null))
    else null
  else null;

{
  resolveSchema: resolveSchema,
  resolvedResponseSchema: resolvedResponseSchema,
  isArraySchema: isArraySchema,
  collectionItemsPath: collectionItemsPath,
  isCollectionResponseSchema: isCollectionResponseSchema,
  arrayItemSchema: arrayItemSchema,
}
