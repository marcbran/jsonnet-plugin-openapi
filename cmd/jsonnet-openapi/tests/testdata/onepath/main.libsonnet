{
  users: {
    get: function(args) std.native('invoke:onepath')('request', [
      {
        method: 'GET',
        path: '/users',
      },
    ]),
  },
}
