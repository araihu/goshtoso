package docspages

import (
	"context"
	"embed"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/expressions"
)

type expressionDemoKey struct{}
type expressionDemoState struct {
	Language string
	Page     int
}

// WithExpressionDemo reads this demo's request-owned language and page selection.
func WithExpressionDemo(r *http.Request) *http.Request {
	language := r.URL.Query().Get("language")
	if language != "pt" && language != "es" && language != "zh" && language != "de" {
		language = "en"
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 || page > 5 {
		page = 2
	}
	return r.WithContext(context.WithValue(r.Context(), expressionDemoKey{}, expressionDemoState{language, page}))
}

func expressionDemoFrom(ctx context.Context) expressionDemoState {
	if state, ok := ctx.Value(expressionDemoKey{}).(expressionDemoState); ok {
		return state
	}
	return expressionDemoState{"en", 2}
}

//go:embed expressions/*.yaml
var demoExpressionFiles embed.FS

var demoExpressionSets = loadDemoExpressionSets()

func loadDemoExpressionSets() map[string]expressions.Set {
	sets := make(map[string]expressions.Set)
	for _, language := range []string{"en", "pt", "es", "zh", "de"} {
		set, err := expressions.LoadFS(demoExpressionFiles, "expressions/"+language+".yaml")
		if err != nil {
			panic(fmt.Errorf("embedded demo expressions: %w", err))
		}
		sets[language] = set
	}
	return sets
}

func translatedDemoSet(language string) expressions.Set {
	return demoExpressionSets[language]
}

func expressionPreview(language string, request bool, child templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		set := translatedDemoSet(language)
		if request {
			return child.Render(expressions.With(ctx, set), w)
		}
		return expressions.NewRenderer(set).Render(ctx, w, child)
	})
}
