package components_test

import (
	"testing"

	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/appshell"
	"github.com/araihu/goshtoso/components/emptystate"
	"github.com/araihu/goshtoso/components/pageheader"
	"github.com/araihu/goshtoso/components/panel"
	"github.com/araihu/goshtoso/components/skeleton"
	"github.com/araihu/goshtoso/components/toolbar"
	"github.com/stretchr/testify/require"
)

func compositionRenderables() map[components.Kind]components.Component {
	return map[components.Kind]components.Component{
		components.KindAppShell:   appshell.AppShell(appshell.Config{}),
		components.KindPageHeader: pageheader.PageHeader(pageheader.Config{}),
		components.KindToolbar:    toolbar.Toolbar(toolbar.Config{}),
		components.KindPanel:      panel.Panel(panel.Config{}),
		components.KindEmptyState: emptystate.EmptyState(emptystate.Config{}),
		components.KindSkeleton:   skeleton.Skeleton(skeleton.Config{}),
	}
}

func TestCompositionRenderablesExposeKinds(t *testing.T) {
	values := compositionRenderables()
	require.Len(t, values, 6)
	for want, value := range values {
		require.Equal(t, want, value.Kind())
	}
}
