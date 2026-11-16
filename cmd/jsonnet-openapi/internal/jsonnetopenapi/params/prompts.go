package params

const paramsPrompt = `Read input.json.

Classify every query parameter of the OpenAPI list endpoint at sourcePath by the role it plays when listing items.

Return only JSON matching the provided schema.

Roles:
- "filter": narrows which items are returned, for example a status, owner, type, date range, list of ids, or free-text search query.
- "shaping": changes what each item or the response contains without changing which items are returned, for example including or expanding related resources, selecting fields, toggling a total count, or choosing a representation.
- "pagination": controls which page is returned or how many items it holds, for example limit, offset, page, page size, cursors, and page tokens.
- "sort": controls the order of the items, for example a sort key or a sort direction.
- "other": anything that fits none of the above, for example an API version, locale, or debugging flag.

Rules:
- Classify every parameter in params exactly once, using its name as given.
- Use the description, schema type, format, and enum values, not only the name.
- A parameter that narrows by time, such as since or until, is a "filter", unless its description says it is a cursor for paging.
- When a parameter both narrows and reshapes the response, prefer "filter".
- Keep reasons short.
`
