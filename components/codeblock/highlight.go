package codeblock

import (
	"bytes"
	htmlpkg "html"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Render returns a Chroma-highlighted <pre><code>…</code></pre> fragment as
// raw HTML, suitable for injection via templ.Raw.
//
// Token spans use the "ch-" class prefix; css/codeblock.css maps a small
// subset of those classes to a 4-color palette. Unknown languages fall back
// to plain (escaped) text inside <pre><code>.
func Render(code, lang string) string {
	lexer := pickLexer(lang)
	if lexer == nil {
		return plainPre(code)
	}
	lexer = chroma.Coalesce(lexer)

	iter, err := lexer.Tokenise(nil, code)
	if err != nil {
		return plainPre(code)
	}

	formatter := chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.ClassPrefix("ch-"),
		chromahtml.WithLineNumbers(false),
	)

	var buf bytes.Buffer
	if err := formatter.Format(&buf, styles.Fallback, iter); err != nil {
		return plainPre(code)
	}
	return buf.String()
}

func pickLexer(lang string) chroma.Lexer {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "templ":
		// No native templ lexer in Chroma; Go gives the closest useful
		// coloring for the templ snippets shown in demos today.
		return lexers.Get("go")
	case "":
		return lexers.Fallback
	default:
		if l := lexers.Get(lang); l != nil {
			return l
		}
		return lexers.Fallback
	}
}

func plainPre(code string) string {
	return `<pre class="chroma"><code>` + htmlpkg.EscapeString(code) + `</code></pre>`
}
