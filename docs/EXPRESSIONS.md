# Customizing built-in expressions

Goshtoso accepts application-owned text for built-in controls, empty states,
status messages, and accessible names. English remains the default. Goshtoso
does not choose languages or translate application content. Your application can
obtain expressions from code, a translation service, a database, or user and
system configuration.

## Configure application defaults once

```go
import "github.com/araihu/goshtoso/expressions"

renderer := expressions.NewRenderer(expressions.Set{
    Pagination: expressions.Pagination{
        PreviousLabel:  "Anterior",
        PreviousAriaLabel: "Página anterior",
        NextLabel:      "Próxima",
        NextAriaLabel:     "Próxima página",
    },
    CodeBlock: expressions.CodeBlock{
        CopyLabel:   "Copiar",
        CopiedLabel: "Copiado!",
        ErrorText:   "Não foi possível copiar",
    },
})

// Reuse the renderer in HTTP handlers, background rendering, or static generation.
err := renderer.Render(r.Context(), w, Page())
```

A zero renderer uses English. Sets are copied by value and contain strings and
functions, not mutable maps. Treat any state captured by a message function as
immutable or synchronize access yourself. Configuration changes can replace an
application's renderer or stored set using the application's normal concurrency
controls; there is no process-global Goshtoso setter.

## Override per request

Choose an expression set using your application's language or user preferences.
Pass it through the request context, including handlers serving HTMX fragments:

```go
ctx := expressions.With(r.Context(), userExpressions)
err := renderer.Render(ctx, w, Page())
```

Middleware can use `r.WithContext(ctx)`. Ordinary `component.Render(ctx, w)` also
honors context expressions; a configured renderer is only necessary when you
want a reusable layer of application defaults. Repeated `expressions.With`
calls overlay nonempty fields, so system, tenant, and user configuration can be
combined without constructing a complete set at each level.

The configured renderer's defaults apply only to renders through that renderer.
A built-in HTTP handler that calls `Render` directly needs system defaults in
its request context too: `expressions.With(expressions.With(r.Context(),
systemExpressions), userExpressions)`. An instance's overrides are not persisted
or transmitted to later HTTP requests.

## Override one component

Components that own built-in expressions expose a typed, copy-returning method:

```go
control := pagination.Pagination(pagination.Config{
    CurrentPage: 2,
    TotalPages:  10,
}).WithExpressions(expressions.Pagination{
    NextLabel: "Continue",
})
```

The result retains its concrete component type and `Kind`. The existing standalone
`navbar.SecondaryRow` helper keeps its return type and uses
`SecondaryConfig.AriaLabel` for its per-instance label override. Existing explicit
config fields still win, such as `search.Config.Label`, `EmptyText`, and
`Placeholder`. Custom content and component slots remain application-owned.

Precedence, independently for each expression:

1. An existing explicit component config field, where one exists.
2. The nearest component's `WithExpressions` value.
3. Request/render-context expressions.
4. Configured renderer defaults.
5. Built-in English.

Component expressions are scoped to the component's render subtree; nested
components inherit them. A nested component can supply its own overrides, and
siblings outside the subtree are unaffected. An empty string or nil function
means inherit; it does not suppress an accessible label or hide a control.
Use the component's visibility/content options when hiding content is intended.

## Complete dynamic messages

Supply functions for messages with parameters. The application owns word order,
number formatting, and plural rules:

```go
set := expressions.Set{
    Combobox: expressions.Combobox{
        SelectedLabel: func(count int) string {
            return translator.SelectedCount(count)
        },
    },
    Pagination: expressions.Pagination{
        PageAriaLabel: func(page int) string {
            return translator.PageAriaLabel(page)
        },
    },
}
```

Carousel and client Combobox render their possible count labels into escaped
HTML data attributes. Browser interactions select those prepared messages;
application functions do not run in JavaScript. Count functions must support
zero as well as positive counts. The client Combobox prepares labels through
the available options plus initially selected values. When changing available
options, render the component again so the prepared messages stay in sync.
Restored client selections are deduplicated and restricted to available options
or values explicitly rendered as initially selected.
CodeBlock copy success/failure, mobile navigation, and file-selection feedback
also use server-resolved text after browser interaction.

Changing render-context expressions affects the next render. It does not
retroactively translate DOM already displayed in a browser. A language change
should render the page or affected fragments again.

For JSON, YAML, and embedded files, see [Expression files](EXPRESSION_FILES.md),
including validation rules and the complete JSON Schema.

## Scope

The [expression inventory](EXPRESSION_INVENTORY.md) maps the public fields to
component source. Product titles, navigation items, table cells, form validation
messages, and other caller-supplied content are not translated automatically.
HTML values, URLs, CSS tokens, diagnostic errors, and raw numeric control values
retain their existing semantics. This API does not implement RTL layout or
locale negotiation.
