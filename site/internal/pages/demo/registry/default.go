// Package registry constructs and indexes the demo site's routable pages.
package registry

import (
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	accordionpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/accordion"
	actiongrouppage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/actiongroup"
	alertpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/alert"
	appshellpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/appshell"
	avatarpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/avatar"
	badgepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/badge"
	bannerpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/banner"
	breadcrumbspage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/breadcrumbs"
	buttonpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/button"
	cardpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/card"
	carouselpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/carousel"
	chatbubblepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/chatbubble"
	checkboxpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/checkbox"
	codeblockpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/codeblock"
	comboboxpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/combobox"
	drawerpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/drawer"
	dropdownpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/dropdown"
	emptystatepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/emptystate"
	fileinputpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/fileinput"
	formpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/form"
	headpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/head"
	iconpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/icon"
	kbdpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/kbd"
	linkpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/link"
	modalpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/modal"
	navbarpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/navbar"
	pageheaderpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/pageheader"
	paginationpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/pagination"
	palettepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/palette"
	panelpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/panel"
	radiopage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/radio"
	rangepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/range"
	ratingpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/rating"
	schemaformpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/schemaform"
	scrollregionpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/scrollregion"
	searchpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/search"
	selectpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/select"
	sidebarpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/sidebar"
	skeletonpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/skeleton"
	spinnerpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/spinner"
	stepspage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/steps"
	structuredinputpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/structuredinput"
	tablepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/table"
	tabspage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/tabs"
	tagslistpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/tagslist"
	textareapage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/textarea"
	textinputpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/textinput"
	toastpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/toast"
	togglepage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/toggle"
	toolbarpage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/toolbar"
	tooltippage "github.com/araihu/goshtoso/site/internal/pages/demo/componentpages/tooltip"
	docspages "github.com/araihu/goshtoso/site/internal/pages/demo/contentpages/docs"
	legalpages "github.com/araihu/goshtoso/site/internal/pages/demo/contentpages/legal"
	modulespages "github.com/araihu/goshtoso/site/internal/pages/demo/contentpages/modules"
	startpages "github.com/araihu/goshtoso/site/internal/pages/demo/contentpages/start"
	chatpage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/chat"
	expensepage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/expense"
	indexpage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/index"
	logspage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/logs"
	profilepage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/profile"
	tickerpage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/ticker"
	todopage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/todo"
	wizardpage "github.com/araihu/goshtoso/site/internal/pages/demo/examplepages/wizard"
)

var defaultPages = mustDefault()

func mustDefault() *Registry {
	definitions := []demo.PageDefinition{
		accordionpage.Definition,
		actiongrouppage.Definition,
		alertpage.Definition,
		appshellpage.Definition,
		avatarpage.Definition,
		badgepage.Definition,
		bannerpage.Definition,
		breadcrumbspage.Definition,
		buttonpage.Definition,
		cardpage.Definition,
		carouselpage.Definition,
		chatbubblepage.Definition,
		checkboxpage.Definition,
		codeblockpage.Definition,
		comboboxpage.Definition,
		drawerpage.Definition,
		dropdownpage.Definition,
		emptystatepage.Definition,
		fileinputpage.Definition,
		formpage.Definition,
		headpage.Definition,
		iconpage.Definition,
		kbdpage.Definition,
		linkpage.Definition,
		modalpage.Definition,
		navbarpage.Definition,
		pageheaderpage.Definition,
		paginationpage.Definition,
		palettepage.Definition,
		panelpage.Definition,
		radiopage.Definition,
		rangepage.Definition,
		ratingpage.Definition,
		schemaformpage.Definition,
		searchpage.Definition,
		selectpage.Definition,
		sidebarpage.Definition,
		skeletonpage.Definition,
		scrollregionpage.Definition,
		spinnerpage.Definition,
		stepspage.Definition,
		structuredinputpage.Definition,
		tablepage.Definition,
		tabspage.Definition,
		tagslistpage.Definition,
		textareapage.Definition,
		textinputpage.Definition,
		toastpage.Definition,
		togglepage.Definition,
		toolbarpage.Definition,
		tooltippage.Definition,
	}
	definitions = append(definitions, docspages.Definitions...)
	definitions = append(definitions, legalpages.Definitions...)
	definitions = append(definitions, modulespages.Definitions...)
	definitions = append(definitions, startpages.Definitions...)
	definitions = append(definitions,
		chatpage.Definition,
		expensepage.Definition,
		logspage.Definition,
		profilepage.Definition,
		tickerpage.Definition,
		todopage.Definition,
		wizardpage.Definition,
		indexpage.Definition,
	)

	pages, err := New(definitions, catalog.ComponentPages())
	if err != nil {
		panic(err)
	}
	return pages
}

// Lookup returns a page from the default site registry.
func Lookup(key string) (demo.PageDefinition, bool) {
	return defaultPages.Lookup(key)
}

// MetaForKey returns crawler metadata from the default site registry.
func MetaForKey(key string) demo.PageMeta {
	return defaultPages.MetaForKey(key)
}

// AllPublicMeta returns home metadata and all default registered pages.
func AllPublicMeta() []demo.PageMeta {
	return defaultPages.AllPublicMeta()
}
