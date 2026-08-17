package registry

import (
	"testing"

	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	popoverpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/popover"
	splitbuttonpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/splitbutton"
	"github.com/stretchr/testify/require"
)

func TestPopoverAndSplitButtonPagesAreCatalogedAndRoutable(t *testing.T) {
	for _, test := range []struct {
		key  string
		kind components.Kind
		page string
	}{
		{key: "components/popover", kind: components.KindPopover, page: popoverpage.Definition.Key},
		{key: "components/splitbutton", kind: components.KindSplitButton, page: splitbuttonpage.Definition.Key},
	} {
		t.Run(test.key, func(t *testing.T) {
			entry, ok := catalog.Lookup(test.key)
			require.True(t, ok)
			require.Equal(t, []components.Kind{test.kind}, entry.Kinds)
			require.Equal(t, test.key, test.page)

			definition, ok := Lookup(test.key)
			require.True(t, ok)
			require.NotNil(t, definition.Content)
		})
	}
}
