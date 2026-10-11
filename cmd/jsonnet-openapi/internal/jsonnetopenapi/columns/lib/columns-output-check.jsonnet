local schema = import 'inference-schema.libsonnet';

local keyFor(column) = std.get(column, 'key', std.join('.', column.path));

local isScalarItem(itemSchema) =
  std.get(itemSchema, 'type', null) != 'object' && std.length(schema.pathProperties(itemSchema)) == 0;

local resolve(itemSchema, path) =
  local parent = schema.schemaAtPath({}, itemSchema, path[0:std.length(path) - 1]);
  {
    schema: if parent.schema == null then null else std.get(schema.pathProperties(parent.schema), path[std.length(path) - 1], null),
    crossedArray: parent.crossedArray,
  };

local columnProblems(itemSchema, column) =
  local path = std.manifestJsonMinified(column.path);
  if std.length(column.path) == 0 then
    if isScalarItem(itemSchema) then []
    else [std.format('column path %s is only valid when the list items are scalar values', [path])]
  else
    local info = resolve(itemSchema, column.path);
    if info.schema == null then [std.format('column path %s does not exist in itemSchema', [path])]
    else if info.crossedArray then [std.format('column path %s crosses an array; a column must resolve to a single value per list item', [path])]
    else [];

function(input, output)
  local checked = [
    { column: column, problems: columnProblems(input.itemSchema, column) }
    for column in output.columns
  ];
  local valid = [entry.column for entry in checked if std.length(entry.problems) == 0];
  local keys = [keyFor(column) for column in valid];
  local duplicates = std.set([key for key in keys if std.count(keys, key) > 1]);
  local unique = [
    column
    for i in std.range(0, std.length(valid) - 1)
    for column in [valid[i]]
    if std.count(keys[0:i], keyFor(column)) == 0
  ];
  local primaries = std.length([column for column in unique if column.priority == 'primary']);
  {
    valid: output {
      sourcePath: input.sourcePath,
      operationId: input.operationId,
      array: input.array,
      columns: unique,
    },
    problems: std.flattenArrays([entry.problems for entry in checked]) + [
      std.format('column key %s is used more than once; keys must be unique', [key])
      for key in duplicates
    ] + [
      std.format('array %s must be %s, as given in input.json', [std.manifestJsonMinified(output.array), std.manifestJsonMinified(input.array)])
      for _ in [null]
      if output.array != input.array
    ] + [
      std.format('exactly one column must have priority "primary", found %d', [primaries])
      for _ in [null]
      if primaries != 1
    ],
  }
