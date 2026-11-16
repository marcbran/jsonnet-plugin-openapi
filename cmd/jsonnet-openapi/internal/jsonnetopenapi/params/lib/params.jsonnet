function(spec, inferred)
  [
    {
      sourcePath: item.sourcePath,
      operationId: std.get(item, 'operationId', null),
      params: [
        { name: param.name, role: param.role }
        for param in item.params
      ],
    }
    for item in inferred
  ]
