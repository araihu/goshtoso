# templ syntax reference

Verbatim from templ.guide. templ v0.3.x.

## File structure

A `.templ` file is Go: package declaration, imports, then a mix of ordinary Go and `templ` blocks.

```templ
package main

import "fmt"
import "time"

var greeting = "Welcome!"   // ordinary Go, outside templ blocks

templ headerTemplate(name string) {
	<header>
		<h1>{ name }</h1>
		<h2>"{ greeting }" comes from ordinary Go code</h2>
	</header>
}
```

## Elements

- **All tags must close**: closing tag `</a>` or self-close `<hr/>`. templ does not infer closure.
- templ knows **void elements** and omits the slash in output: `<img src="x"/>` → `<img src="x">`, `<br/>` → `<br>`.
- Elements and attributes can contain `{ }` expressions.

```templ
templ component() {
	<div>Test</div>
	<img src="images/test.png"/>
	<br/>
}
```
Output: `<div>Test</div><img src="images/test.png"><br>`

## Expressions (`{ }`)

Interpolation outputs **context-aware HTML-escaped** text. Supported: strings, numbers (`int`/`uint`/`float`/`complex`), booleans, anything `fmt.Stringer`, and functions returning a value or `(value, error)`. A returned error aborts `Render` with location info.

```templ
templ component() {
	<div>{ "print this" }</div>
	<div>{ `and this` }</div>
	<div>Number of the day: { 1 }</div>
	<div>{ strings.ToUpper("abc") }</div>
	<div>{ getString() }</div>   // func returning (string, error)
}
```

Escaping is automatic — this is rendered inert, not executed:
```templ
<div>{ `</div><script>alert('hello!')</script><div>` }</div>
```

## Control flow

Standard Go statements, used directly inside markup, no escaping.

```templ
// if / else
templ login(isLoggedIn bool) {
	if isLoggedIn {
		<div>Welcome back!</div>
	} else {
		<input name="login" type="button" value="Log in"/>
	}
}

// switch
templ userTypeDisplay(userType string) {
	switch userType {
		case "test":  <span>{ "Test user" }</span>
		case "admin": <span>{ "Admin user" }</span>
		default:      <span>{ "Unknown user" }</span>
	}
}

// for
templ nameList(items []Item) {
	<ul>
		for _, item := range items {
			<li>{ item.Name }</li>
		}
	</ul>
}
```

**Text starting with `if`/`switch`/`for`** is parsed as a statement. To emit it as literal text, use a string expression (`{ "if ..." }`) or capitalize the word.

## Attributes

### Constant / string-expression
```templ
<button value={ name }>{ content }</button>
```

### Conditional attribute
```templ
<hr style="padding: 10px"
	if true {
		class="itIsTrue"
	}
/>
```

### Boolean attribute
Presence = true, absence = false. With a variable use `?=`:
```templ
<hr noshade/>
<hr noshade?={ false } />
```

### Spread attributes (`templ.Attributes` = `map[string]any`)
```templ
templ component(shouldBeUsed bool, attrs templ.Attributes) {
	<p { attrs... }>Text</p>
	<hr
		if shouldBeUsed {
			{ attrs... }
		}
	/>
}
```

### Attribute key expression
```templ
templ paragraph(testID string) {
	<p { "data-" + testID }="paragraph">Text</p>
}
```

### URL attributes
`href` is auto-sanitized (JavaScript URLs stripped). Constants are not sanitized.
```templ
<a href={ p.URL }>{ strings.ToUpper(p.Name) }</a>          // sanitized
<a href={ templ.URL(urlString) }>...</a>                   // explicit sanitize
<a href={ templ.SafeURL(myURL) }>...</a>                   // bypass sanitize (trusted)
```

### JSON attribute (e.g. web component, Alpine)
```templ
<div x-data={ templ.JSONString(data) }>...</div>
```

### JavaScript handler attribute
`on*` handlers expect a `script` template reference or `templ.JSFuncCall`:
```templ
script withParameters(a string, b string, c int) {
	console.log(a, b, c);
}
templ Button(text string) {
	<button onClick={ withParameters("test", text, 123) }>{ text }</button>
}
```

## CSS class & style attributes

```templ
<button class={ className }>{ text }</button>              // single
<button class={ "button", className }>{ text }</button>    // multiple
<button class={ "button", templ.KV("is-primary", isPrimary) }>{ text }</button>  // conditional
<button class={ map[string]bool{"active": isActive, "disabled": isDisabled} }>x</button>
```

Style attribute (constant/expr — see security.md for restrictions):
```templ
<button style={ style }>{ text }</button>
<button style={ style1, style2 }>{ text }</button>
```

`css` blocks generate a unique class name:
```templ
css primaryClassName() {
	background-color: #ffffff;
	color: { red };
}
css loading(percent int) {
	width: { fmt.Sprintf("%d%%", percent) };
}
templ button(text string, isPrimary bool) {
	<button class={ templ.KV(primaryClassName(), isPrimary) }>{ text }</button>
}
```

Supported style types: `string`, `templ.SafeCSS`, `map[string]string`, `templ.KeyValue[string,string]`, `templ.KeyValue[string,bool]` (conditional inclusion).

## Template composition

```templ
templ showAll() {
	@left()
	@middle()
	@right()
}
templ left() { <div>Left</div> }
```

### Children — `{ children... }`
```templ
templ wrapChildren() {
	<div id="wrapper">
		{ children... }
	</div>
}
// caller:
@wrapChildren() {
	<p>nested content goes into the slot</p>
}
```
In Go: `templ.WithChildren(ctx, contents)`, `templ.GetChildren(ctx)`, `templ.ClearChildren(ctx)`.

### Components as parameters (type `templ.Component`)
```templ
templ layout(contents templ.Component) {
	<div id="contents">
		@contents
	</div>
}
templ root() {
	@layout(paragraph("Dynamic contents"))
}
```

### Joining
```templ
@templ.Join(component1(), component2())
```

### Sharing across packages
Same package: directly accessible. Different package: capitalize to export, import the package, call `@components.Hello()`.

## Rendering raw HTML

Bypasses all escaping — trusted source only.
```templ
@templ.Raw("<div>Hello, World!</div>")
```

## Raw Go

Declare variables/functions outside `templ` blocks. Inside, use `{{ }}` for statements:
```templ
templ component() {
	{{ name := "world" }}
	<div>{ name }</div>
}
```

## Render once (`OnceHandle`)

Render shared content (e.g. a `<script>`) only once per context even if the component renders many times.
```templ
var helloHandle = templ.NewOnceHandle()

templ hello(name string) {
	@helloHandle.Once() {
		<script>
			function hello(name) { alert('Hello, ' + name + '!'); }
		</script>
	}
	<div onclick="hello('{ name }')">Click</div>
}
```

## Context

`context.Context` flows through `Render`. Use it for request-scoped values (theme, auth, i18n) and for children/once handling. Pass via HTTP middleware. Access in a component with a code block reading `ctx`. (See templ.guide/syntax-and-usage/context for middleware patterns.)

## Comments

```templ
<!-- HTML comment — appears in output -->
// Go comment — stripped, not rendered
```
