{
  items: {
    get: function(args) std.native('invoke:multiparam')('request', [
      local query = std.get(args, 'query', {});
      local headers = std.get(args, 'headers', {});
      {
        method: 'GET',
        path: '/items',
        query: {
          zebra: std.get(query, 'zebra', null),
          alpha: std.get(query, 'alpha', null),
        },
        headers: {
          HeaderB: std.get(headers, 'HeaderB', null),
          HeaderA: std.get(headers, 'HeaderA', null),
        },
      },
    ]),
  },
}
