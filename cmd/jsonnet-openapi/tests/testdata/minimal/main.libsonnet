{
  users: {
    get: function(args) std.native('invoke:minimal')('request', [
      {
        method: 'GET',
        path: '/users',
      },
    ]),
    userId(userId): {
      get: function(args) std.native('invoke:minimal')('request', [
        {
          method: 'GET',
          path: std.format('/users/%s', [std.toString(userId)]),
        },
      ]),
    },
  },
}
