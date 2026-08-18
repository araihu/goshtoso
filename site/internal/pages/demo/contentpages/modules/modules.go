package modulespages

import "github.com/a-h/templ"

// ChartsModuleContent is the site-owned landing page for the Charts module.
// The server uses this content instead of the charts package's documentation
// page so the module landing can include the site-level showcase frames.
func ChartsModuleContent() templ.Component {
	return chartsModuleContent()
}
