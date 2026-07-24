package components_test

import (
	"testing"

	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/alert"
	"github.com/araihu/goshtoso/components/breadcrumbs"
	"github.com/araihu/goshtoso/components/drawer"
	"github.com/araihu/goshtoso/components/dropdown"
	"github.com/araihu/goshtoso/components/link"
	"github.com/araihu/goshtoso/components/modal"
	"github.com/araihu/goshtoso/components/navbar"
	"github.com/araihu/goshtoso/components/pagination"
	"github.com/araihu/goshtoso/components/sidebar"
	"github.com/araihu/goshtoso/components/spinner"
	"github.com/araihu/goshtoso/components/steps"
	"github.com/araihu/goshtoso/components/tabs"
	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/components/tooltip"
	"github.com/stretchr/testify/require"
)

func feedbackRenderables() map[components.Kind]components.Component {
	return map[components.Kind]components.Component{
		components.KindAlert:           alert.Alert(alert.Config{}),
		components.KindModal:           modal.Modal(modal.Config{}),
		components.KindAlertDialog:     modal.AlertDialog(modal.AlertDialogConfig{}),
		components.KindToastContainer:  toast.ToastContainer(toast.ContainerConfig{}),
		components.KindToast:           toast.Toast(toast.Config{}),
		components.KindMessageToast:    toast.MessageToast(toast.MessageConfig{}),
		components.KindOOBToast:        toast.OOBToast(toast.Config{}),
		components.KindOOBMessageToast: toast.OOBMessageToast(toast.MessageConfig{}),
		components.KindDrawer:          drawer.Drawer(drawer.Config{}),
		components.KindSpinner:         spinner.Spinner(spinner.Config{}),
		components.KindSteps:           steps.Steps(steps.Config{}),
		components.KindTooltip:         tooltip.Tooltip("", ""),
	}
}

func navigationRenderables() map[components.Kind]components.Component {
	return map[components.Kind]components.Component{
		components.KindBreadcrumbs:    breadcrumbs.Breadcrumbs(breadcrumbs.Config{}),
		components.KindDropdown:       dropdown.Dropdown(dropdown.Config{}),
		components.KindLink:           link.Link(""),
		components.KindNavbar:         navbar.Navbar(navbar.Config{}),
		components.KindPagination:     pagination.Pagination(pagination.Config{}),
		components.KindSidebar:        sidebar.Sidebar(sidebar.Config{}),
		components.KindSidebarOverlay: sidebar.Overlay(sidebar.OverlayConfig{}),
		components.KindTabs:           tabs.Tabs(tabs.Config{}),
	}
}

func TestFeedbackRenderablesExposeKinds(t *testing.T) {
	values := feedbackRenderables()
	require.Len(t, values, 12)
	for want, value := range values {
		require.Equal(t, want, value.Kind())
	}
}

func TestNavigationRenderablesExposeKinds(t *testing.T) {
	values := navigationRenderables()
	require.Len(t, values, 8)
	for want, value := range values {
		require.Equal(t, want, value.Kind())
	}
}
