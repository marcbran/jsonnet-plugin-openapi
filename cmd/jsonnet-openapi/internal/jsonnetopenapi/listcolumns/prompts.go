package listcolumns

const columnsPrompt = `Read input.json.

Choose useful table-view columns for the OpenAPI list response at sourcePath.

Return only JSON matching the provided schema.

Rules:
- Choose roughly 3 to 6 columns.
- Column paths must be property paths relative to the list item, for example ["id"] or ["owner", "name"].
- Do not invent properties that are not supported by itemSchema.
- Prefer human-readable names or titles, stable identifiers, status fields, timestamps, small counts, and concise scalar fields.
- Avoid descriptions, URLs, blobs, nested arrays, and large object values unless they are clearly useful as table columns.
- Use priority "primary" for the main identifying column.
- Prefer a human-readable name, title, label, slug, username, or email column as "primary" when available.
- Use a stable identifier as "primary" only when no human-readable identifying column is available.
- Use priority "secondary" for generally useful columns and "tertiary" for lower-value columns.
- Use kind values that describe display intent, such as "identifier", "name", "status", "timestamp", "count", "text", "number", or "boolean".
- Keep labels short and human-readable.
`
