package listdetaillinks

const listDetailPrompt = `Read input.json.

Infer whether the OpenAPI list response at sourcePath has a canonical GET detail path among detailPaths.

Return only JSON matching the provided schema.

Rules:
- Choose "detail_elsewhere" only when there is a clear canonical detail GET endpoint for the list item type.
- Choose "no_detail_get" when the list item is an event, stats/summary object, search result, relationship record, activity feed item, or otherwise has no canonical detail GET path in detailPaths.
- Choose "uncertain" when there is not enough evidence.
- For "detail_elsewhere", targetPath must be one path from detailPaths and array must be the array path from input.json.
- For "no_detail_get" or "uncertain", targetPath must be null.
- Do not invent paths.
- Do not infer variable mappings.
`

const varsPrompt = `Read input.json.

Infer JSON property paths on the array item that provide values for the target path params listed in missingParams.

Return only JSON matching the provided schema.

Rules:
- Only infer vars for params in missingParams.
- Each vars value must be a property path relative to the array item, for example ["account", "id"] or ["name"].
- Do not include params that are already present in inheritedParams.
- Do not invent properties that are not supported by itemSchema.
- Match the meaning of each target path param, not just its name.
- Prefer stable canonical identifiers over display names.
- Prefer exact or clearly equivalent property names when available, for example an "id" param from an "id" property, or a "name" param from a "name" property.
- For params ending in "_id" or named "id", prefer stable id-like fields over names, slugs, titles, or URLs.
- For params ending in "_name" or named "name", prefer stable name-like fields over display titles or descriptions.
- For slug/key/code params, prefer slug/key/code fields over human-readable labels.
- Avoid URLs, descriptions, titles, summaries, display names, timestamps, booleans, counts, and status fields unless the target param clearly asks for that value.
- Return vars as an array of objects with param and path fields.
- If a missing param cannot be resolved from itemSchema, omit it from vars and explain that in reason.
`
