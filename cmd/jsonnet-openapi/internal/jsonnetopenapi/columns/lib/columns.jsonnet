local keyFor(path) = std.join('.', path);

function(spec, inferred)
  [
    {
      sourcePath: item.sourcePath,
      operationId: std.get(item, 'operationId', null),
      array: item.array,
      columns: [
        {
          key: std.get(column, 'key', keyFor(column.path)),
          path: column.path,
          label: column.label,
          kind: column.kind,
          priority: column.priority,
        }
        for column in item.columns
      ],
    }
    for item in inferred
  ]
