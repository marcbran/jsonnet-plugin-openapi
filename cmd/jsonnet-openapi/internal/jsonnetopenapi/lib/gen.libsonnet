function(api)
  local j = import 'jsonnet/main.libsonnet';

  local le = j.Fodder.LineEnd();

  local isJsonnetIdent(s) =
    if std.length(s) == 0 then false
    else
      local len = std.length(s);
      local isAsciiLetter(c) =
        (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z');
      local isAsciiDigit(c) = c >= '0' && c <= '9';
      local identStart(c) = c == '_' || isAsciiLetter(c);
      local identPart(c) = identStart(c) || isAsciiDigit(c);
      local check(i) =
        if i >= len then true
        else if i == 0 then identStart(s[i]) && check(i + 1)
        else identPart(s[i]) && check(i + 1);
      check(0);

  local jsonnetKeywords = [
    'assert',
    'else',
    'error',
    'false',
    'for',
    'function',
    'if',
    'import',
    'importstr',
    'in',
    'local',
    'null',
    'self',
    'super',
    'tailstrict',
    'then',
    'true',
  ];

  local isJsonnetKeyword(s) = std.member(jsonnetKeywords, s);

  local isUnquotedFieldName(s) = isJsonnetIdent(s) && !isJsonnetKeyword(s);

  local objectField(name, expr) =
    if std.type(name) == 'string' && isUnquotedFieldName(name) then j.Field(name, expr) else j.Field(j.String(name), expr);

  local pathParamInner(seg) =
    local len = std.length(seg);
    if len >= 2 && std.substr(seg, 0, 1) == '{' && std.substr(seg, len - 1, 1) == '}' then
      std.substr(seg, 1, len - 2)
    else null;

  local mangledPathVar(name) =
    if isJsonnetIdent(name) && !isJsonnetKeyword(name) then name
    else 'p_' + std.md5(name);

  local hasLeaf(node) = std.get(node, 'leaf', null) != null;
  local childrenOf(node) = std.get(node, 'children', {});
  local hasChildren(node) =
    local ch = childrenOf(node);
    std.length(std.objectFields(ch)) > 0;

  local pathExpr(op) =
    local fmt = std.get(op, 'pathFormat', '/');
    local ns = std.get(op, 'pathArgNames', []);
    if std.length(ns) == 0 then j.String(fmt)
    else j.Std.format(
      j.String(fmt),
      j.Array([
        j.Std.toString(j.Var(mangledPathVar(n)))
        for n in ns
      ])
    );

  local emptyObject = j.Object([]);

  local argField(bucketVar, p) =
    local bucketExpr = j.Var(bucketVar);
    objectField(
      p.name,
      if p.required then j.Index(bucketExpr, j.String(p.name))
      else j.Std.get(bucketExpr, j.String(p.name)).default(j.Null)
    );

  local paramObject(bucketVar, params) =
    if std.length(params) == 0 then
      emptyObject
    else
      j.Object([argField(bucketVar, p) for p in params]).closeFodder(le);

  local bucketBind(bucketKey, varName) =
    j.LocalBind(
      varName,
      j.Std.get(j.Var('args'), j.String(bucketKey)).default(emptyObject)
    );

  local inputObject(op) =
    local q = std.get(op, 'queryParams', []);
    local h = std.get(op, 'headerParams', []);
    local base = [
      j.Field('method', j.String('GET')),
      j.Field('path', pathExpr(op)),
    ];
    local withQuery =
      if std.length(q) > 0 then base + [j.Field('query', paramObject('query', q))] else base;
    local withHeaders =
      if std.length(h) > 0 then withQuery + [j.Field('headers', paramObject('headers', h))] else withQuery;
    local obj = j.Object(withHeaders).closeFodder(le);
    local queryB = if std.length(q) > 0 then bucketBind('query', 'query') else null;
    local headersB = if std.length(h) > 0 then bucketBind('headers', 'headers') else null;
    if queryB != null && headersB != null then
      j.Local(queryB, j.Local(headersB, obj.fodder(le)).fodder(le))
    else if queryB != null then
      j.Local(queryB, obj.fodder(le))
    else if headersB != null then
      j.Local(headersB, obj.fodder(le))
    else
      obj;

  local opFunction(op) =
    j.Function(
      [j.Parameter('args')],
      j.Apply(
        j.Apply(
          j.Member(j.Var('std'), 'native'),
          [j.String('invoke:' + api.service)]
        ),
        [j.String('request'), j.Array([inputObject(op)]).closeFodder(le)]
      )
    );

  local leafAsGet(node) =
    j.Object([j.Field('get', opFunction(node.leaf))]).closeFodder(le);

  local trieField(k, expr) =
    local inner = pathParamInner(k);
    if inner != null then
      local pv = mangledPathVar(inner);
      local fieldId = if isUnquotedFieldName(inner) then inner else j.String(inner);
      j.FieldFunction(fieldId, [j.Parameter(pv)], expr)
    else
      objectField(k, expr);

  local trieToExpr(node) =
    if hasLeaf(node) && !hasChildren(node) then leafAsGet(node)
    else if hasChildren(node) then
      local ch = childrenOf(node);
      local otherKeys = [k for k in std.sort(std.objectFields(ch)) if k != '_'];
      local underscoreField =
        if std.objectHas(ch, '_') then
          [objectField(
            if std.objectHas(ch, 'get') then '_get' else 'get',
            opFunction(ch['_'].leaf)
          )]
        else [];
      j.Object(underscoreField + [trieField(k, trieToExpr(ch[k])) for k in otherKeys]).closeFodder(le)
    else error 'invalid trie node';

  local pkg(meta) =
    j.Local(
      j.LocalBind('p', j.Import('pkg/main.libsonnet')),
      j.Apply(j.Member(j.Var('p'), 'pkg'), [
        j.Object([
          j.Field('repo', j.String(meta.pkgRepo)),
          j.Field('branch', j.String('openapi/%s' % meta.service)),
          j.Field('path', j.String('openapi-%s' % meta.service)),
          j.Field('target', j.String(meta.service)),
        ]).closeFodder(le),
        j.String('%s %s' % [std.get(meta.info, 'title', ''), std.get(meta.info, 'version', '')]),
      ]).fodder(le)
    );

  {
    'main.libsonnet': j.manifestJsonnet(trieToExpr(api.trie)),
    'pkg.libsonnet': j.manifestJsonnet(pkg(api)),
  }
