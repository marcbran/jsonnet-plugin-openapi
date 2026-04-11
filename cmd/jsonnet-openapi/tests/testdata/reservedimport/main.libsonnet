{
  'import': {
    get: function(args) std.native('invoke:reservedimport')('request', [
      {
        method: 'GET',
        path: '/import',
      },
    ]),
  },
}
