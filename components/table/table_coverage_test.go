package table

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// renderT renders a templ component to a string for substring assertions.
func renderT(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

func mustContainAll(t *testing.T, html string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(html, w) {
			t.Fatalf("missing %q in:\n%s", w, html)
		}
	}
}

func mustNotContain(t *testing.T, html string, unwanted ...string) {
	t.Helper()
	for _, u := range unwanted {
		if strings.Contains(html, u) {
			t.Fatalf("unexpected %q in:\n%s", u, html)
		}
	}
}

// --- Pure class / helper coverage -------------------------------------------

func TestBadgeCellClasses_AllColors(t *testing.T) {
	cases := map[string]string{
		"success":   "bg-success text-on-success",
		"danger":    "bg-danger text-on-danger",
		"warning":   "bg-warning text-on-warning",
		"info":      "bg-info text-on-info",
		"primary":   "bg-primary text-on-primary",
		"secondary": "bg-secondary text-on-secondary",
		"neutral":   "bg-surface-alt text-on-surface",
		"unknown":   "border border-outline",
		"":          "border border-outline",
	}
	for color, want := range cases {
		got := BadgeCellClasses(color)
		if !strings.Contains(got, "inline-flex") {
			t.Fatalf("BadgeCellClasses(%q) missing base; got %q", color, got)
		}
		if !strings.Contains(got, want) {
			t.Fatalf("BadgeCellClasses(%q) = %q; want substring %q", color, got, want)
		}
	}
}

func TestStatusBadgeClasses_Branches(t *testing.T) {
	cases := map[string]string{
		"active":   "bg-success text-on-success",
		"success":  "bg-success text-on-success",
		"canceled": "bg-danger text-on-danger",
		"danger":   "bg-danger text-on-danger",
		"pending":  "", // default: base only, no color
	}
	for status, want := range cases {
		got := StatusBadgeClasses(status)
		if !strings.Contains(got, "inline-flex") {
			t.Fatalf("StatusBadgeClasses(%q) missing base; got %q", status, got)
		}
		if want != "" && !strings.Contains(got, want) {
			t.Fatalf("StatusBadgeClasses(%q) = %q; want %q", status, got, want)
		}
		if want == "" && (strings.Contains(got, "bg-success") || strings.Contains(got, "bg-danger")) {
			t.Fatalf("StatusBadgeClasses(%q) leaked a color class: %q", status, got)
		}
	}
}

func TestContainerClasses_RootClassAppended(t *testing.T) {
	base := Config{}.ContainerClasses()
	mustContainAll(t, base, "overflow-x-auto", "overflow-y-clip", "border-outline")
	withRoot := Config{RootClass: "shadow-lg"}.ContainerClasses()
	mustContainAll(t, withRoot, "overflow-x-auto", "shadow-lg")
}

func TestColCount_Combinations(t *testing.T) {
	cols := []Column{{Key: "a"}, {Key: "b"}}
	if n := (Config{Columns: cols}).ColCount(); n != 2 {
		t.Fatalf("plain ColCount = %d; want 2", n)
	}
	if n := (Config{Columns: cols, ShowCheckbox: true}).ColCount(); n != 3 {
		t.Fatalf("checkbox ColCount = %d; want 3", n)
	}
	withActions := Config{
		Columns: cols,
		Rows:    []Row{{ID: "1", Actions: templ.Raw("x")}},
	}
	if n := withActions.ColCount(); n != 3 {
		t.Fatalf("actions ColCount = %d; want 3", n)
	}
	withExpand := Config{
		Columns:      cols,
		ShowCheckbox: true,
		Rows:         []Row{{ID: "1", Expandable: true}},
	}
	if n := withExpand.ColCount(); n != 4 {
		t.Fatalf("checkbox+expand ColCount = %d; want 4", n)
	}
}

func TestNextSortDir_Cycle(t *testing.T) {
	cfg := Config{}
	if got := cfg.NextSortDir("name"); got != SortAsc {
		t.Fatalf("unsorted -> %q; want asc", got)
	}
	cfg = Config{SortBy: "name", SortDir: SortAsc}
	if got := cfg.NextSortDir("name"); got != SortDesc {
		t.Fatalf("asc -> %q; want desc", got)
	}
	cfg = Config{SortBy: "name", SortDir: SortDesc}
	if got := cfg.NextSortDir("name"); got != SortNone {
		t.Fatalf("desc -> %q; want none", got)
	}
	cfg = Config{SortBy: "other", SortDir: SortAsc}
	if got := cfg.NextSortDir("name"); got != SortAsc {
		t.Fatalf("different column -> %q; want asc", got)
	}
}

func TestPaginationPages(t *testing.T) {
	var nilP *PaginationConfig
	if got := nilP.PaginationPages(); got != nil {
		t.Fatalf("nil pagination pages = %v; want nil", got)
	}
	if got := (&PaginationConfig{TotalPages: 0}).PaginationPages(); got != nil {
		t.Fatalf("zero total pages = %v; want nil", got)
	}
	got := (&PaginationConfig{TotalPages: 3}).PaginationPages()
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Fatalf("PaginationPages(3) = %v; want [1 2 3]", got)
	}
}

func TestPaginationConfigPointerHelpers(t *testing.T) {
	var nilP *PaginationConfig
	if nilP.NextPage() != 2 {
		t.Fatalf("nil NextPage = %d; want 2", nilP.NextPage())
	}
	if nilP.GetContainerHeight() != "400px" {
		t.Fatalf("nil GetContainerHeight = %q; want 400px", nilP.GetContainerHeight())
	}
	if nilP.IsInfiniteScroll() || nilP.IsContained() {
		t.Fatal("nil pagination should not be infinite/contained")
	}
	p := &PaginationConfig{Mode: PaginationInfiniteScroll, CurrentPage: 4, ContainerHeight: "60vh"}
	if p.NextPage() != 5 {
		t.Fatalf("NextPage = %d; want 5", p.NextPage())
	}
	if p.GetContainerHeight() != "60vh" {
		t.Fatalf("GetContainerHeight = %q; want 60vh", p.GetContainerHeight())
	}
	if !p.IsInfiniteScroll() || !p.IsContained() {
		t.Fatal("expected infinite + contained")
	}
	// Infinite but no container height -> not contained (Pattern B).
	pb := &PaginationConfig{Mode: PaginationInfiniteScroll}
	if pb.IsContained() {
		t.Fatal("infinite without height should not be contained")
	}
}

func TestRowClickableRoleAndActionable(t *testing.T) {
	if r := (Row{Link: "/x"}); r.ClickableRole() != "link" || !r.IsActionable() {
		t.Fatal("link row should be link/actionable")
	}
	if r := (Row{OnClick: "go()"}); r.ClickableRole() != "button" || !r.IsActionable() {
		t.Fatal("onclick row should be button/actionable")
	}
	htmxRow := Row{HTMX: &RowHTMXConfig{Get: "/x"}}
	if htmxRow.ClickableRole() != "button" || !htmxRow.HasHTMXAction() {
		t.Fatal("htmx-get row should be button with htmx action")
	}
	postRow := Row{HTMX: &RowHTMXConfig{Post: "/x"}}
	if !postRow.HasHTMXAction() {
		t.Fatal("htmx-post row should have htmx action")
	}
	// Actions-only / expandable-only keep default role (nested controls own focus).
	if r := (Row{Actions: templ.Raw("a")}); r.ClickableRole() != "" || !r.IsActionable() {
		t.Fatal("actions-only row: empty role but actionable")
	}
	if r := (Row{Expandable: true}); r.ClickableRole() != "" || !r.IsActionable() {
		t.Fatal("expandable-only row: empty role but actionable")
	}
	if r := (Row{}); r.ClickableRole() != "" || r.IsActionable() {
		t.Fatal("plain row: empty role, not actionable")
	}
	emptyHTMX := Row{HTMX: &RowHTMXConfig{}}
	if emptyHTMX.HasHTMXAction() {
		t.Fatal("empty HTMX config should not count as action")
	}
}

func TestConfigPredicates(t *testing.T) {
	cfg := Config{
		Columns: []Column{{Key: "a", Sortable: true}, {Key: "b"}},
		Rows: []Row{
			{ID: "1", Link: "/1"},
			{ID: "2", Expandable: true},
			{ID: "3", Actions: templ.Raw("x")},
		},
	}
	if !cfg.HasLinkedRows() {
		t.Fatal("want HasLinkedRows")
	}
	if !cfg.HasActionableRows() {
		t.Fatal("want HasActionableRows")
	}
	if !cfg.HasExpandableRows() {
		t.Fatal("want HasExpandableRows")
	}
	if !cfg.HasActions() {
		t.Fatal("want HasActions")
	}
	if !cfg.HasSortableColumns() {
		t.Fatal("want HasSortableColumns")
	}
	if cfg.HasFilters() {
		t.Fatal("no filters configured")
	}
	if !cfg.IsSortedBy("") && cfg.IsSortedBy("a") {
		t.Fatal("IsSortedBy false for unsorted column a")
	}
	empty := Config{}
	if empty.HasLinkedRows() || empty.HasActionableRows() || empty.HasExpandableRows() ||
		empty.HasActions() || empty.HasSortableColumns() {
		t.Fatal("empty config should report no features")
	}
}

func TestIDHelpersAndDefaults(t *testing.T) {
	cfg := Config{}
	if cfg.GetID() != "table" {
		t.Fatalf("default id = %q", cfg.GetID())
	}
	if cfg.TbodyID() != "table-tbody" || cfg.TheadID() != "table-thead" ||
		cfg.PaginationID() != "table-pagination" || cfg.FilterBarID() != "table-filters" {
		t.Fatal("derived ids wrong for default")
	}
	named := Config{ID: "users"}
	if named.TbodyID() != "users-tbody" {
		t.Fatalf("named tbody = %q", named.TbodyID())
	}
	if cfg.LazyLoadTrigger() != "load" {
		t.Fatalf("default lazy trigger = %q", cfg.LazyLoadTrigger())
	}
	if (Config{LazyTrigger: "click from:#go"}).LazyLoadTrigger() != "click from:#go" {
		t.Fatal("custom lazy trigger lost")
	}
}

func TestHTMXValueAccessors(t *testing.T) {
	if (Config{}).HTMXEndpointValue() != "" || (Config{}).HTMXTargetValue() != "" {
		t.Fatal("nil HTMX should yield empty endpoint/target")
	}
	cfg := Config{HTMX: &HTMXConfig{Endpoint: "/api", Target: "#t"}}
	if cfg.HTMXEndpointValue() != "/api" || cfg.HTMXTargetValue() != "#t" {
		t.Fatal("HTMX accessors wrong")
	}
}

func TestColumnAlignmentClasses(t *testing.T) {
	left := ColumnCellClasses(Column{Width: "w-32"})
	mustContainAll(t, left, "p-4", "w-32")
	mustNotContain(t, left, "text-center", "text-right")
	center := ColumnCellClasses(Column{Align: "center"})
	mustContainAll(t, center, "text-center")
	right := ColumnHeaderClasses(Column{Align: "right", Width: "w-10"})
	mustContainAll(t, right, "text-right", "w-10")
}

func TestRowAndSortableHeaderClasses(t *testing.T) {
	if (Config{}).RowClasses() != "" {
		t.Fatal("default variant rows should have no extra class")
	}
	striped := Config{Appearance: AppearanceStriped}.RowClasses()
	mustContainAll(t, striped, "odd:bg-surface-alt")

	cfg := Config{SortBy: "name", SortDir: SortAsc}
	sorted := cfg.SortableHeaderClasses("name")
	mustContainAll(t, sorted, "cursor-pointer", "text-primary")
	unsorted := cfg.SortableHeaderClasses("other")
	mustNotContain(t, unsorted, "text-primary dark:text-primary-dark")

	mustContainAll(t, (Config{}).CheckboxClasses(), "appearance-none")
	mustContainAll(t, (Config{}).TableClasses(), "w-full")
	mustContainAll(t, (Config{}).TheadClasses(), "border-b")
	mustContainAll(t, (Config{}).TbodyClasses(), "divide-y")
	mustContainAll(t, (Config{}).CellClasses(), "p-4")
	mustContainAll(t, (Config{}).HeaderCellClasses(), "p-4")
	mustContainAll(t, ActionButtonClasses(), "text-primary")
}

func TestHyphenToCamel(t *testing.T) {
	cases := map[string]string{
		"filtered-table": "filteredTable",
		"a-b-c":          "aBC",
		"plain":          "plain",
		"-leading":       "Leading",
		"trailing-":      "trailing",
		"with-9digit":    "with9digit",
	}
	for in, want := range cases {
		if got := hyphenToCamel(in); got != want {
			t.Fatalf("hyphenToCamel(%q) = %q; want %q", in, got, want)
		}
	}
}

// --- Render coverage for exported templ entry points ------------------------

func TestRenderDefaultTable(t *testing.T) {
	html := renderT(t, Table(Config{
		Columns: []Column{{Key: "name", Label: "Name"}},
		Rows:    []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "Alice"}}}},
	}))
	mustContainAll(t, html, "<table", "id=\"table\"", "Name", "Alice", "<tbody")
}

func TestRenderCaptionAndStriped(t *testing.T) {
	html := renderT(t, Table(Config{
		Caption:    "Customer list",
		Appearance: AppearanceStriped,
		Columns:    []Column{{Key: "name", Label: "Name"}},
		Rows:       []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "Bob"}}}},
	}))
	mustContainAll(t, html, "<caption", "Customer list", "odd:bg-surface-alt")
}

func TestRenderCellContentVariants(t *testing.T) {
	html := renderT(t, Table(Config{
		Columns: []Column{
			{Key: "badge", Label: "Badge"},
			{Key: "code", Label: "Code"},
			{Key: "desc", Label: "Desc"},
			{Key: "comp", Label: "Comp"},
			{Key: "plain", Label: "Plain"},
		},
		Rows: []Row{{
			ID: "1",
			Cells: map[string]Cell{
				"badge": {Text: "OK", BadgeColor: "success"},
				"code":  {Text: "go test", Code: true},
				"desc":  {Text: "Alice", Description: "id-42"},
				"comp":  {Component: templ.Raw("<em>custom</em>")},
				"plain": {Text: "plaincell"},
			},
		}},
	}))
	mustContainAll(t, html,
		"bg-success text-on-success", "OK", // badge
		"<code", "go test", // code
		"id-42", "Alice", // description stacked
		"<em>custom</em>", // component
		"plaincell",       // plain text
	)
}

func TestRenderCheckboxHeaderAndCells(t *testing.T) {
	html := renderT(t, Table(Config{
		ShowCheckbox: true,
		Columns:      []Column{{Key: "name", Label: "Name"}},
		Rows:         []Row{{ID: "42", Cells: map[string]Cell{"name": {Text: "Z"}}}},
	}))
	mustContainAll(t, html,
		`x-data="{ checkAll: false }"`,
		`x-model="checkAll"`, // header checkbox
		`id="checkAll"`,      // header id
		`for="user42"`,       // row checkbox label
		`x-bind:checked="checkAll"`,
	)
}

func TestRenderSortableHeaderIcons(t *testing.T) {
	base := Config{
		HTMX:    &HTMXConfig{Endpoint: "/api/components/table/rows"},
		Columns: []Column{{Key: "name", Label: "Name", Sortable: true}},
		Rows:    []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "A"}}}},
	}
	// Unsorted -> dual arrow (opacity-40), sortable header has hx-get.
	unsorted := renderT(t, TableHead(base))
	mustContainAll(t, unsorted, "hx-get=", "opacity-40", "hx-target=\"#table-tbody\"")

	asc := base
	asc.SortBy, asc.SortDir = "name", SortAsc
	mustContainAll(t, renderT(t, TableHead(asc)), "text-primary")

	desc := base
	desc.SortBy, desc.SortDir = "name", SortDesc
	descHTML := renderT(t, TableHead(desc))
	mustContainAll(t, descHTML, "M10 3a.75.75 0 01.75.75") // desc arrow path

	// Sortable column WITHOUT an endpoint falls back to a plain header.
	noEndpoint := Config{Columns: []Column{{Key: "name", Label: "Name", Sortable: true}}}
	plain := renderT(t, TableHead(noEndpoint))
	mustNotContain(t, plain, "hx-get=")
}

func TestRenderTableHeadContentNoWrapper(t *testing.T) {
	html := renderT(t, TableHeadContent(Config{
		Columns: []Column{{Key: "name", Label: "Name"}},
	}))
	mustContainAll(t, html, "<tr>", "Name")
	mustNotContain(t, html, "<thead")
}

func TestRenderRowLinkModes(t *testing.T) {
	cols := []Column{{Key: "name", Label: "Name"}}
	cell := map[string]Cell{"name": {Text: "Row"}}

	spa := renderT(t, TableRow(Config{Columns: cols}, Row{ID: "1", Link: "/p/1", Cells: cell}))
	mustContainAll(t, spa, `hx-get="/p/1"`, `hx-target="#main-content-area"`, `role="link"`, `tabindex="0"`, "onkeydown")

	boost := renderT(t, TableRow(Config{Columns: cols}, Row{ID: "2", Link: "/p/2", LinkMode: LinkBoost, Cells: cell}))
	mustContainAll(t, boost, `hx-select="body"`, `hx-target="body"`)

	full := renderT(t, TableRow(Config{Columns: cols}, Row{ID: "3", Link: "/p/3", LinkMode: LinkFull, Cells: cell}))
	mustContainAll(t, full, "window.location.href=", "onauxclick")
	mustNotContain(t, full, "hx-get=")
}

func TestRenderRowActionAttrs(t *testing.T) {
	cols := []Column{{Key: "name", Label: "Name"}}
	cell := map[string]Cell{"name": {Text: "R"}}

	onclick := renderT(t, TableRow(Config{Columns: cols}, Row{ID: "1", OnClick: "openModal()", Cells: cell}))
	mustContainAll(t, onclick, `onclick="openModal()"`, `role="button"`)

	get := renderT(t, TableRow(Config{Columns: cols}, Row{
		ID:    "2",
		HTMX:  &RowHTMXConfig{Get: "/row/2", Target: "#slot", PushURL: true},
		Cells: cell,
	}))
	mustContainAll(t, get, `hx-get="/row/2"`, `hx-target="#slot"`, `hx-push-url="true"`, `hx-swap="innerHTML"`)

	post := renderT(t, TableRow(Config{Columns: cols}, Row{
		ID:    "3",
		HTMX:  &RowHTMXConfig{Post: "/row/3", Target: "#slot", Swap: "outerHTML"},
		Cells: cell,
	}))
	mustContainAll(t, post, `hx-post="/row/3"`, `hx-swap="outerHTML"`)
}

func TestRenderRowAlpineAttrs(t *testing.T) {
	html := renderT(t, TableRow(Config{Columns: []Column{{Key: "name", Label: "N"}}}, Row{
		ID:          "1",
		AlpineAttrs: map[string]string{"x-show": "visible"},
		Cells:       map[string]Cell{"name": {Text: "A"}},
	}))
	mustContainAll(t, html, `x-show="visible"`)
}

func TestRenderExpandableRowWithDetail(t *testing.T) {
	row := Row{
		ID:         "7",
		Expandable: true,
		Detail:     templ.Raw("<p>detail body</p>"),
		Cells:      map[string]Cell{"name": {Text: "Exp"}},
	}
	cfg := Config{Columns: []Column{{Key: "name", Label: "N"}}, Rows: []Row{row}}
	html := renderT(t, TableRow(cfg, row))
	// templ escapes single quotes to &#39; inside attribute values.
	mustContainAll(t, html,
		"openRows[&#39;7&#39;] = !openRows[&#39;7&#39;]", // toggle click
		"x-bind:class", // chevron rotate
		"detail body",  // detail panel
		`x-show="openRows[&#39;7&#39;]"`,
	)
}

func TestRenderRowWithActionsColumn(t *testing.T) {
	row := Row{
		ID:      "1",
		Actions: templ.Raw(`<button>Edit</button>`),
		Cells:   map[string]Cell{"name": {Text: "A"}},
	}
	cfg := Config{Columns: []Column{{Key: "name", Label: "N"}}, Rows: []Row{row}}
	html := renderT(t, TableRow(cfg, row))
	mustContainAll(t, html, "<button>Edit</button>", "justify-end")
}

func TestRenderLazyLoadTbody(t *testing.T) {
	html := renderT(t, Table(Config{
		LazyLoad: true,
		HTMX:     &HTMXConfig{Endpoint: "/api/components/table/rows"},
		Columns:  []Column{{Key: "name", Label: "Name"}},
	}))
	mustContainAll(t, html,
		`hx-get="/api/components/table/rows"`,
		`hx-trigger="load"`,
		"Loading...", // loadingIndicator
		"animate-spin",
	)
}

func TestRenderInfiniteScrollSentinel(t *testing.T) {
	cfg := Config{
		HTMX:    &HTMXConfig{Endpoint: "/api/components/table/rows"},
		Columns: []Column{{Key: "name", Label: "Name"}},
		Rows:    []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "A"}}}},
		Pagination: &PaginationConfig{
			Mode:        PaginationInfiniteScroll,
			CurrentPage: 1,
			PerPage:     10,
			HasMore:     true,
		},
	}
	body := renderT(t, TableBody(cfg))
	mustContainAll(t, body,
		`id="table-sentinel"`,
		"data-hx-get=",
		"IntersectionObserver",
		"variant=infinite",
	)
	// TableRows (append response) also emits the sentinel without a tbody.
	rows := renderT(t, TableRows(cfg))
	mustContainAll(t, rows, `id="table-sentinel"`)
	mustNotContain(t, rows, "<tbody")
}

func TestRenderLegacyInfiniteScrollSentinel(t *testing.T) {
	cfg := Config{
		HTMX:           &HTMXConfig{Endpoint: "/api/components/table/rows"},
		Columns:        []Column{{Key: "name", Label: "Name"}},
		Rows:           []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "A"}}}},
		InfiniteScroll: &InfiniteScrollConfig{NextPage: 2, HasMore: true},
	}
	body := renderT(t, TableBody(cfg))
	mustContainAll(t, body, `id="table-sentinel"`, "page=2")
}

func TestRenderPaginationControls(t *testing.T) {
	cfg := Config{
		HTMX:       &HTMXConfig{Endpoint: "/api/components/table/rows"},
		Columns:    []Column{{Key: "name", Label: "Name"}},
		Rows:       []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "A"}}}},
		Pagination: &PaginationConfig{CurrentPage: 2, TotalPages: 4, PerPage: 3},
	}
	full := renderT(t, Table(cfg))
	mustContainAll(t, full, "Page 2 of 4", `id="table-pagination"`)

	nav := renderT(t, TablePaginationNav(cfg))
	mustContainAll(t, nav, "nav") // pagination component renders a nav element
	mustNotContain(t, nav, "Page 2 of 4")

	// Single page -> no pagination chrome.
	single := Config{
		Columns:    []Column{{Key: "name", Label: "N"}},
		Rows:       []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "A"}}}},
		Pagination: &PaginationConfig{CurrentPage: 1, TotalPages: 1, PerPage: 3},
	}
	mustNotContain(t, renderT(t, TablePagination(single)), "table-pagination")
}

func TestRenderContainedInfiniteScrollVariants(t *testing.T) {
	mk := func(checkbox, expand bool) Config {
		rows := []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "A"}}, Expandable: expand}}
		return Config{
			HTMX:         &HTMXConfig{Endpoint: "/api/components/table/rows"},
			Columns:      []Column{{Key: "name", Label: "N"}},
			Rows:         rows,
			ShowCheckbox: checkbox,
			Pagination: &PaginationConfig{
				Mode:            PaginationInfiniteScroll,
				CurrentPage:     1,
				ContainerHeight: "300px",
				HasMore:         false,
			},
		}
	}
	// checkbox + expandable
	mustContainAll(t, renderT(t, Table(mk(true, true))),
		"max-height: 300px", `x-data="{ checkAll: false, openRows: {} }"`)
	// checkbox only
	mustContainAll(t, renderT(t, Table(mk(true, false))),
		"max-height: 300px", `x-data="{ checkAll: false }"`)
	// expandable only
	mustContainAll(t, renderT(t, Table(mk(false, true))),
		"max-height: 300px", `x-data="{ openRows: {} }"`)
	// plain contained
	mustContainAll(t, renderT(t, Table(mk(false, false))), "max-height: 300px; overflow-y: auto;")
}

func TestRenderNonContainedContainerVariants(t *testing.T) {
	cols := []Column{{Key: "name", Label: "N"}}
	cell := map[string]Cell{"name": {Text: "A"}}

	// Non-contained checkbox + expandable rows.
	both := renderT(t, Table(Config{
		ShowCheckbox: true,
		Columns:      cols,
		Rows:         []Row{{ID: "1", Cells: cell, Expandable: true}},
	}))
	mustContainAll(t, both, `x-data="{ checkAll: false, openRows: {} }"`)
	mustNotContain(t, both, "max-height")

	// Non-contained expandable-only rows.
	expandOnly := renderT(t, Table(Config{
		Columns: cols,
		Rows:    []Row{{ID: "1", Cells: cell, Expandable: true}},
	}))
	mustContainAll(t, expandOnly, `x-data="{ openRows: {} }"`)
	mustNotContain(t, expandOnly, "max-height")
}

func TestRenderFilterBarVariants(t *testing.T) {
	memberships := []FilterOption{{Value: "vip", Label: "VIP"}, {Value: "std", Label: "Standard"}}
	cfg := Config{
		ID:      "filtered",
		HTMX:    &HTMXConfig{Endpoint: "/api/components/table/rows"},
		Columns: []Column{{Key: "name", Label: "Name"}},
		Rows:    []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "A"}}}},
		Filters: &FilterConfig{
			Collapsible:       true,
			InitiallyExpanded: false,
			Filters: []Filter{
				{Key: "search", Label: "Search", Type: FilterSearch, Placeholder: "Find..."},
				{Key: "membership", Label: "Membership", Type: FilterSelect, Options: memberships},
				{Key: "active", Label: "Active only", Type: FilterToggle},
				{Key: "team", Label: "Team", Type: FilterSelect, OptionsHTMX: &FilterOptionsHTMXConfig{Get: "/api/teams"}},
			},
		},
	}
	html := renderT(t, Table(cfg))
	mustContainAll(t, html,
		"<script", "filteredFilters", // filterScript + Alpine.data name
		`x-data="filteredFilters"`,   // wrapper
		`id="filtered-filters"`,      // filter bar id
		`@click="filtersExpanded`,    // collapsible toggle
		`type="search"`,              // search input
		`@input.debounce.300ms`,      // search trigger
		"<select", "VIP", "Standard", // static select options
		`@change="applyFilters()"`,          // select/toggle change
		`type="checkbox"`,                   // toggle input
		"Active only",                       // toggle label
		`hx-get="/api/teams"`, "Loading...", // dynamic select options
	)
}

func TestRenderInlineFilterVariant(t *testing.T) {
	cfg := Config{
		ID:      "inline",
		HTMX:    &HTMXConfig{Endpoint: "/api/x"},
		Columns: []Column{{Key: "name", Label: "Name"}},
		Rows:    []Row{{ID: "1", Cells: map[string]Cell{"name": {Text: "A"}}}},
		Filters: &FilterConfig{
			Appearance: FilterAppearanceInline,
			Filters:    []Filter{{Key: "q", Type: FilterSearch, Placeholder: "Search…"}},
		},
	}
	html := renderT(t, Table(cfg))
	mustContainAll(t, html, `id="inline-filters"`, "flex flex-wrap items-end gap-3", `type="search"`)
	// Inline variant drops the collapsible toggle.
	mustNotContain(t, html, `@click="filtersExpanded`)
}

func TestRenderStandaloneActionButtonAndStatusBadge(t *testing.T) {
	mustContainAll(t, renderT(t, ActionButton("Edit")), "<button", "Edit", "text-primary")
	mustContainAll(t, renderT(t, StatusBadge("Active", "active")), "Active", "bg-success")
	mustContainAll(t, renderT(t, StatusBadge("Pending", "pending")), "Pending", "inline-flex")
}

func TestRenderImageCell(t *testing.T) {
	withImg := renderT(t, ImageCell("/avatar.png", "Alice", "alice@example.com"))
	mustContainAll(t, withImg, "Alice", "alice@example.com")
	// No image URL -> avatar falls back to initials.
	noImg := renderT(t, ImageCell("", "Bob Jones", "team"))
	mustContainAll(t, noImg, "Bob Jones", "team")
}
