local responseRefName(ref) = std.strReplace(ref, '#/components/responses/', '');
local schemaRefName(ref) = std.strReplace(ref, '#/components/schemas/', '');

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

local schemaIdentity(schema) =
  local ref = refString(schema);
  if ref != null then ref else std.get(schema, 'x-jsonnet-openapi-ref', null);

local normalizeSchema(spec, schema, seen=[]) =
  local ref = refString(schema);
  if ref != null then
    if std.member(seen, ref) then {
      'x-jsonnet-openapi-ref': ref,
      'x-jsonnet-openapi-recursiveRef': true,
    } else
      normalizeSchema(spec, spec.components.schemas[schemaRefName(ref)], seen + [ref]) {
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
  if ref != null then normalizeSchema(spec, spec.components.schemas[schemaRefName(ref)], [ref]) {
    'x-jsonnet-openapi-ref': ref,
  } else normalizeSchema(spec, schema);

local resolvedResponseSchema(spec, response) =
  resolveSchema(spec, responseSchema(spec, response));

local isArraySchema(schema) =
  std.type(schema) == 'object' &&
  std.get(schema, 'type', null) == 'array';

local isArrayOfRefs(schema) =
  isArraySchema(schema) &&
  schemaIdentity(std.get(schema, 'items', null)) != null;

local hasArrayRefProperty(schema) =
  std.type(schema) == 'object' &&
  std.get(schema, 'type', null) == 'object' &&
  std.objectHas(schema, 'properties') &&
  std.any([
    isArrayOfRefs(schema.properties[property])
    for property in std.objectFields(schema.properties)
  ]);

local isListResponseSchema(schema) =
  isArrayOfRefs(schema) || hasArrayRefProperty(schema);

local responseArrays(spec, schema) =
  if isArraySchema(schema) then [
    {
      array: [],
      itemSchema: resolveSchema(spec, std.get(schema, 'items', null)),
    },
  ]
  else if std.type(schema) == 'object' &&
          std.get(schema, 'type', null) == 'object' &&
          std.objectHas(schema, 'properties') then [
    {
      array: [property],
      itemSchema: resolveSchema(spec, std.get(schema.properties[property], 'items', null)),
    }
    for property in std.objectFields(schema.properties)
    if isArraySchema(schema.properties[property])
  ]
  else [];

local arrayRefs(schema) =
  if isArrayOfRefs(schema) then [
    {
      array: [],
      ref: schemaIdentity(schema.items),
    },
  ]
  else if std.type(schema) == 'object' &&
          std.get(schema, 'type', null) == 'object' &&
          std.objectHas(schema, 'properties') then [
    {
      array: [property],
      ref: schemaIdentity(schema.properties[property].items),
    }
    for property in std.objectFields(schema.properties)
    if isArrayOfRefs(schema.properties[property])
  ]
  else [];

local arrayItemSchema(spec, schema, array) =
  if std.length(array) == 0 then
    resolveSchema(spec, std.get(schema, 'items', null))
  else if std.length(array) == 1 &&
          std.type(schema) == 'object' &&
          std.objectHas(schema, 'properties') &&
          std.objectHas(schema.properties, array[0]) then
    resolveSchema(spec, std.get(schema.properties[array[0]], 'items', null))
  else null;

{
  responseSchema: responseSchema,
  resolveSchema: resolveSchema,
  resolvedResponseSchema: resolvedResponseSchema,
  isArraySchema: isArraySchema,
  isArrayOfRefs: isArrayOfRefs,
  hasArrayRefProperty: hasArrayRefProperty,
  isListResponseSchema: isListResponseSchema,
  responseArrays: responseArrays,
  arrayRefs: arrayRefs,
  arrayItemSchema: arrayItemSchema,
}
