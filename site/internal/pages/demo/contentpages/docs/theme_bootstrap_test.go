package docspages

import (
	"bytes"
	"context"
	"github.com/araihu/goshtoso/components/themeeditor"
	"github.com/araihu/goshtoso/themes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestThemeCatalogCannotInjectMarkup(t *testing.T) {
	definition, err := themes.Preset(themes.KeyGoshtoso)
	require.NoError(t, err)
	var output bytes.Buffer
	require.NoError(t, themeeditor.ThemeEditor(themeeditor.Config{ID: "hostile-data", Initial: definition, Catalog: []themeeditor.CatalogItem{{ID: "evil", Label: `</section><script>window.injected=true</script>`, Definition: definition}}}).Render(context.Background(), &output))
	require.NotContains(t, output.String(), "<script>window.injected=true</script>")
	require.Contains(t, output.String(), "data-theme-editor=")
}
