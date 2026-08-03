package server

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestTableRowsEndpointPreservesVariantAndPerPage(t *testing.T) {
	got := tableRowsEndpoint("sortable", 6)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse table rows endpoint: %v", err)
	}
	if parsed.Path != "/api/components/table/rows" {
		t.Fatalf("path = %q", parsed.Path)
	}
	if parsed.Query().Get("variant") != "sortable" {
		t.Fatalf("variant = %q", parsed.Query().Get("variant"))
	}
	if parsed.Query().Get("per_page") != "6" {
		t.Fatalf("per_page = %q", parsed.Query().Get("per_page"))
	}
}

func TestTableRowsEndpointWithoutVariantUsesBaseEndpoint(t *testing.T) {
	if got := tableRowsEndpoint("", 3); got != "/api/components/table/rows" {
		t.Fatalf("endpoint = %q", got)
	}
}

func TestTableRowsResponseKeepsStablePaginationHost(t *testing.T) {
	render := func(query string) string {
		t.Helper()
		request := httptest.NewRequest("GET", "/api/components/table/rows?"+query, nil)
		response := httptest.NewRecorder()
		new(Server).handleTableRows(response, request)
		return response.Body.String()
	}

	onePage := render("table_id=inline-filtered-table&page=1&per_page=3&search=alice")
	if !strings.Contains(onePage, `id="inline-filtered-table-pagination"`) ||
		!strings.Contains(onePage, `hx-swap-oob="true"`) ||
		!strings.Contains(onePage, ` hidden>`) {
		t.Fatalf("one-page response must replace a hidden pagination host: %s", onePage)
	}
	if strings.Contains(onePage, "Page 1 of") {
		t.Fatalf("one-page response rendered stale pagination chrome: %s", onePage)
	}

	multiPage := render("table_id=inline-filtered-table&page=1&per_page=3")
	if !strings.Contains(multiPage, `id="inline-filtered-table-pagination"`) ||
		!strings.Contains(multiPage, "Page 1 of 4") {
		t.Fatalf("multi-page response must restore pagination controls: %s", multiPage)
	}
	if strings.Contains(multiPage, ` hidden>`) {
		t.Fatalf("multi-page response must show pagination host: %s", multiPage)
	}
}

func TestTableRowsResponsePreservesInlineFilterColumnContract(t *testing.T) {
	query := url.Values{
		"_filter":  {"1"},
		"variant":  {"inline-filtered"},
		"table_id": {"inline-filtered-table"},
		"per_page": {"3"},
	}.Encode()
	request := httptest.NewRequest("GET", "/api/components/table/rows?"+query, nil)
	response := httptest.NewRecorder()
	new(Server).handleTableRows(response, request)
	body := response.Body.String()

	if response.Code != 200 {
		t.Fatalf("status = %d, body = %s", response.Code, body)
	}
	if strings.Contains(body, "order_by=") {
		t.Fatalf("inline filter response introduced sortable headers: %s", body)
	}
	for _, want := range []string{
		`id="inline-filtered-table-thead"`,
		`id="inline-filtered-table-pagination"`,
		`variant=inline-filtered`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("inline filter response missing %q: %s", want, body)
		}
	}
}

func TestTableRowsResponsePreservesCombinedSortFilterState(t *testing.T) {
	query := url.Values{
		"_filter":    {"1"},
		"membership": {"Gold"},
		"order_by":   {"name"},
		"order_dir":  {"asc"},
		"page":       {"1"},
		"per_page":   {"3"},
		"table_id":   {"filtered-table"},
	}.Encode()
	request := httptest.NewRequest("GET", "/api/components/table/rows?"+query, nil)
	response := httptest.NewRecorder()
	new(Server).handleTableRows(response, request)
	body := response.Body.String()

	if response.Code != 200 {
		t.Fatalf("status = %d, body = %s", response.Code, body)
	}
	for _, want := range []string{
		`data-table-sort-by="name"`,
		`data-table-sort-dir="asc"`,
		`membership=Gold&amp;order_by=name&amp;order_dir=desc`,
		`membership=Gold&amp;order_by=name&amp;order_dir=asc&amp;page=2`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("combined-state response missing %q: %s", want, body)
		}
	}
}

func TestTableRowsRejectsUnsafeTableIDAndEscapesValidTargets(t *testing.T) {
	invalid := []string{
		`x"><img src=x onerror=alert(1)>`,
		"table id",
		"table.selector",
		"table#selector",
		"table:selector",
	}
	for _, tableID := range invalid {
		t.Run(tableID, func(t *testing.T) {
			query := url.Values{
				"table_id": {tableID},
				"page":     {"1"},
				"per_page": {"3"},
			}.Encode()
			request := httptest.NewRequest("GET", "/api/components/table/rows?"+query, nil)
			response := httptest.NewRecorder()
			new(Server).handleTableRows(response, request)

			if response.Code != 400 {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), tableID) || strings.Contains(response.Body.String(), "<img") {
				t.Fatalf("unsafe table ID reflected in response: %s", response.Body.String())
			}
		})
	}

	validID := "inline-filtered-table_2"
	query := url.Values{
		"table_id": {validID},
		"page":     {"1"},
		"per_page": {"3"},
	}.Encode()
	request := httptest.NewRequest("GET", "/api/components/table/rows?"+query, nil)
	response := httptest.NewRecorder()
	new(Server).handleTableRows(response, request)
	body := response.Body.String()
	if response.Code != 200 {
		t.Fatalf("status = %d, body = %s", response.Code, body)
	}
	for _, want := range []string{
		`id="inline-filtered-table_2-thead"`,
		`hx-target="#inline-filtered-table_2-tbody"`,
		`id="inline-filtered-table_2-pagination"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("valid table ID response missing %q: %s", want, body)
		}
	}
}
