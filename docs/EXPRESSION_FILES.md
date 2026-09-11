# Expression files

Store built-in component labels and messages in JSON or YAML, then load them as an expressions.Set. Your application chooses which set to use; Goshtoso supplies English text for fields you leave unspecified.

## Load and render

Create locales/pt.yaml with the text you want to replace. Group and property names match the Go API exactly, including capitalization. This file replaces the empty-state title and description, along with three pagination labels.

```yaml
# yaml-language-server: $schema=./expressions.schema.json
EmptyState:
  Title: Ainda não há nada aqui
  Description: Os itens aparecerão aqui quando estiverem disponíveis.
Pagination:
  NextLabel: Próxima
  NextAriaLabel: Próxima página
  PageAriaLabel: Página {page}
```

Save the following program as main.go beside the locales directory. It embeds the YAML file in the executable, loads the expressions, and writes an empty-state component’s HTML to standard output. The rendered title and description come from the file.

```go
package main

import (
    "context"
    "embed"
    "log"
    "os"

    "github.com/araihu/goshtoso/components/emptystate"
    "github.com/araihu/goshtoso/expressions"
)

//go:embed locales/pt.yaml
var locales embed.FS

func main() {
    set, err := expressions.LoadFS(locales, "locales/pt.yaml")
    if err != nil {
        log.Fatal(err)
    }

    renderer := expressions.NewRenderer(set)
    component := emptystate.EmptyState(emptystate.Config{})
    if err := renderer.Render(context.Background(), os.Stdout, component); err != nil {
        log.Fatal(err)
    }
}
```

In a web application, create the renderer at startup and reuse it in your handlers, passing the request context and response writer to Render. To support several languages, load each file once and choose a set for each request.

See the [Internationalization guide](EXPRESSIONS.md) for application defaults, request overrides, and precedence.

## Files and byte slices

LoadFS accepts any fs.FS, including embed.FS and os.DirFS. Use LoadFile when you have a disk path, or ParseJSON and ParseYAML when the contents are already available as []byte. The file loaders select the format from the .json, .yaml, or .yml extension. All four functions return (Set, error); check the error before using the set.

JSON uses the same groups and properties as YAML. For example, these are the pagination overrides from the YAML file above:

```json
{
  "$schema": "./expressions.schema.json",
  "Pagination": {
    "NextLabel": "Próxima",
    "NextAriaLabel": "Próxima página",
    "PageAriaLabel": "Página {page}"
  }
}
```

Omitted fields and empty strings leave existing defaults in place. A loaded set contains only the supplied overrides, so it can also be passed to expressions.With or combined with another set using expressions.Merge.

## Messages with values

Some messages include a value supplied by the component. Pagination.PageAriaLabel uses {page}: “Página {page}” produces “Página 3” for page 3. Each message accepts only the placeholder listed in the reference below. Use {{ and }} to include literal braces.

File messages substitute values as text. For plural forms or locale-specific number formatting, assign a Go callback to the field after loading the set.

## Validation and property reference

The loaders reject unknown or duplicate properties, nulls, non-string expressions, invalid placeholders, and multiple documents. In YAML, quote numbers and booleans when you intend them as text. YAML aliases and merge keys are not supported.

[Download the JSON Schema](../expressions/schema.json) to use it in your editor. Save the schema as expressions.schema.json beside your translation files. The YAML comment and JSON $schema property shown above associate the file with the schema in editors that support it, enabling property completion and validation while you edit. You can also obtain the schema with expressions.JSONSchema(). The loaders validate the input themselves and do not fetch the $schema URL.

The schema lists every supported property, its English default, and any message placeholder. The documentation site displays the same schema as an expandable tree.
