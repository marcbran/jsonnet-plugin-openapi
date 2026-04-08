{
  'user-profiles': {
    get: function(args) std.native('invoke:unsafeident')('request', [
      local headers = std.get(args, 'headers', {});
      {
        method: 'GET',
        path: '/user-profiles',
        headers: {
          'X-Trace-Id': std.get(headers, 'X-Trace-Id', null),
        },
      },
    ]),
  },
}
