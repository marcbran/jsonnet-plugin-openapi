{
  users: {
    userId(userId): {
      get: function(args) std.native('invoke:paramonly')('request', [
        {
          method: 'GET',
          path: std.format('/users/%s', [std.toString(userId)]),
        },
      ]),
    },
  },
}
