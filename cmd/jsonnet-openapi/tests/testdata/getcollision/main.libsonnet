{
  r: {
    _get: function(args) std.native('invoke:getcollision')('request', [
      {
        method: 'GET',
        path: '/r',
      },
    ]),
    get: {
      get: function(args) std.native('invoke:getcollision')('request', [
        {
          method: 'GET',
          path: '/r/get',
        },
      ]),
    },
    id(id): {
      get: function(args) std.native('invoke:getcollision')('request', [
        {
          method: 'GET',
          path: std.format('/r/%s', [std.toString(id)]),
        },
      ]),
    },
  },
}
