# Expression files

Expression files use the same case-sensitive group and field names as the Go API.
Omitted fields and empty strings inherit the next layer of defaults. Loading a
file returns a partial `Set`, ready for `NewRenderer`, `With`, or `Merge`.

```yaml
# yaml-language-server: $schema=./expressions.schema.json
Pagination:
  NextLabel: Próxima
  PageAriaLabel: Página {page}
EmptyState:
  Title: Ainda não há nada aqui
```

Use `expressions.LoadFile("locales/pt.yaml")` for disk files and
`expressions.LoadFS(locales, "pt.yaml")` for an `fs.FS`, including `embed.FS` and
`os.DirFS`. Both choose JSON or YAML from the `.json`, `.yaml`, or `.yml` extension.
When you already have bytes, use `expressions.ParseJSON(data)` or
`expressions.ParseYAML(data)`. Every loader returns `(Set, error)`; handle the
error before using the result. Load shared files once and choose the appropriate
set for each request. The loaders do not watch files or select a language.

For callback fields, use the named placeholder in the schema, such as `{page}`,
`{count}`, or `{label}`. `Página {page}` renders as `Página 3`. Double braces
(`{{` and `}}`) produce literal braces. Argument text is inserted once and never
interpreted as another placeholder or executable code. File messages perform
simple substitution; supply Go callbacks after loading when you need plural
rules or locale-specific number formatting.

Parsing rejects unknown and duplicate properties, nulls, non-string expressions,
invalid placeholders, and multiple documents. YAML values must be strings; quote
numbers and booleans when they are intended as text. YAML aliases and merge keys
are unsupported.

The [JSON Schema](../expressions/schema.json) lists all supported properties,
English defaults, and placeholder rules. Save it beside your translation files
for editor completion and validation. JSON documents can point to it with
`"$schema": "./expressions.schema.json"`; YAML editors supporting schema
associations can use the comment above. `expressions.JSONSchema()` returns a copy
of the same schema bytes. The parser never fetches a `$schema` URL.

The schema is generated from the public Go fields and their documentation with
`go generate ./expressions`. A drift test keeps it aligned with the API. The
Internationalization site's SchemaTree reads that same schema.


See [the Internationalization guide](EXPRESSIONS.md) for defaults and override precedence.
