{
  items: {
    get: function(args) std.native('invoke:multiparam')('request', [
      local query = std.get(args, 'query', {});
      local headers = std.get(args, 'headers', {});
      {
        method: 'GET',
        path: '/items',
        query: {
          alpha: std.get(query, 'alpha', null),
          zebra: std.get(query, 'zebra', null),
        },
        headers: {
          HeaderA: std.get(headers, 'HeaderA', null),
          HeaderB: std.get(headers, 'HeaderB', null),
        },
      },
    ]),
  },
}
