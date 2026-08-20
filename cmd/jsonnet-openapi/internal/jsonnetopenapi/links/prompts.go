package links

const linksFromResourcesPrompt = `Read input.json.

Identify every distinct place in the resource response at sourcePath that links to another resource's detail GET endpoint from detailPaths.

Return only JSON matching the provided schema: an object with a "links" array, empty if none found.

Rules:
- "at" is a property path, relative to the response root, to where this link is anchored. If "at" crosses an array property, the link occurs once per item of that array. Use [] for "at" when the link is anchored directly at the response root.
- The schema at "at" must be an object, UNLESS "at" crosses an array whose items are plain identifier values (strings or numbers, not objects) — for example an "acknowledged_user_ids" array of raw ids. In that case "at" MUST include that array property as its own last segment, for example "at": ["acknowledged_user_ids"]; each item of the array is then one link, and the item itself is the id. An array of scalar ids must never be left out of "at" and referenced instead through a "path" elsewhere (in "keys" or in a target param) — a "path" is always evaluated relative to the object at "at" and must never itself cross an array; only "at" may cross an array. Never let "at" end at a scalar when it did not cross an array to get there — a single scalar field is not a valid anchor by itself; if the value you want is a top-level scalar field with no array involved, use "at": [] instead and reach the field through "keys"/target params.
- "keys" locates this link within the merged links output, as a list of segments. Each segment is either {"const": "<label>", "path": null} for a fixed label or {"const": null, "path": [...]} for a property path, relative to the object at "at", that produces a distinguishing value. Include a trailing "path" segment whenever "at" crosses an array, to distinguish entries; omit it for links that occur only once. When "at" crosses an array of scalar ids (per the rule above), use {"const": null, "path": []} for that trailing segment, since the item itself, not a sub-property, is the distinguishing value.
- Choose "const" labels that name the relationship or the target resource, never a raw identifier field name: prefer "service" over "service_id", "owner" over "owner_id", "plan" over "plan_id". A reader of the merged output should see a category, not something that looks like it already holds a value. If "at" crosses a property whose own name is already a good category name (like "members", "addresses", or "acknowledged_user_ids"), reuse it, adapted into a readable label (for example "acknowledged users").
- Prefer stable identifiers over display names for "path" segments; unlike "const" labels, "path" segments are expected to resolve to id-like values, since they distinguish entries dynamically rather than label them.
- "targetPath" must be exactly one path from detailPaths. Do not invent paths.
- Only report links you are confident about; skip fields that merely resemble identifiers without pointing to one of detailPaths' resource types.
- A resource can have zero, one, or many links; enumerate all that you find as separate entries, even when several entries share the same "targetPath" — this is expected, for example when a resource references several distinct users in different roles.
- Do not infer variable mappings for target path params; that happens separately.
`

const linksFromCollectionsPrompt = `Read input.json.

The response at sourcePath is a list endpoint. Identify the array of items it returns and whether each item has a canonical detail GET endpoint among detailPaths.

Return only JSON matching the provided schema: an object with a "links" array. Emit at most one link for the list's primary item array; return an empty array when the items have no canonical detail GET.

Rules:
- "at" is a property path, relative to the response root, to the item array. If the response body is itself the array (the root is an array), use "at": []. Otherwise "at" is the path of the array property, for example ["data"] or ["items"]. "at" always crosses this array, so the link occurs once per item.
- Only emit a link when the array items are the resource's own records and have a clear canonical detail GET endpoint in detailPaths. Return an empty "links" array when the list items are events, stats/summary objects, search results, relationship records, activity-feed entries, or otherwise have no canonical detail GET path in detailPaths.
- "keys" locates each item's link within the merged output. Because "at" crosses the item array, "keys" MUST end in a {"const": null, "path": [...]} segment whose path, relative to the array item, resolves to the item's stable identifier (for example ["id"]), so entries stay distinct. Do not add a leading const label unless it is needed to disambiguate.
- "targetPath" must be exactly one path from detailPaths. Do not invent paths.
- Prefer stable id-like fields over display names, slugs, titles, or URLs for "path" segments.
- Do not infer variable mappings for target path params; that happens separately.
`

const linkVarsPrompt = `Read input.json.

Infer JSON property paths, relative to the object at "at", that provide values for the target path params listed in missingParams.

Return only JSON matching the provided schema.

Rules:
- Only infer vars for params in missingParams.
- Each vars value must be a property path relative to the object at "at", for example ["account", "id"] or ["name"].
- If itemSchema is itself a scalar value (only possible when "at" crosses an array of scalar ids), use an empty path [] to mean the value is used directly.
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
- Echo sourcePath, targetPath, at, and keys exactly as given in input.json.
- If a missing param cannot be resolved from itemSchema, omit it from vars and explain that in reason.
`
