package examples

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExamplesIndexPublishesEveryRegisteredAppAsExtractableRecipe(t *testing.T) {
	var rendered strings.Builder
	require.NoError(t, IndexContent().Render(context.Background(), &rendered))
	body := rendered.String()

	for _, example := range []struct {
		path   string
		source string
	}{
		{path: "/examples/todo", source: "todo.templ"},
		{path: "/examples/expense", source: "expense.templ"},
		{path: "/examples/chat", source: "chat.templ"},
		{path: "/examples/logs", source: "logs.templ"},
		{path: "/examples/profile", source: "profile.templ"},
		{path: "/examples/ticker", source: "ticker.templ"},
		{path: "/examples/wizard", source: "wizard.templ"},
	} {
		require.Contains(t, body, `href="`+example.path+`"`)
		require.Contains(t, body, "site/internal/pages/demo/examples/"+example.source)
	}

	for _, label := range []string{"Components", "States", "Complexity", "View source"} {
		require.Contains(t, body, label)
	}
	require.Equal(t, 7, strings.Count(body, `data-example-recipe`))
	require.Equal(t, 7, strings.Count(body, `/assets/images/homepage/examples/`))
	require.Contains(t, body, "Run example")
}
