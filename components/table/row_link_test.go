package table

import "testing"

func TestRowLinkAttrsUseInertFullNavigationData(t *testing.T) {
	link := "/people/o'connor?x=1&next=\\home"
	attrs := rowLinkAttrs(Row{
		Link:     link,
		LinkMode: LinkFull,
	})

	if got := attrs["data-table-row-link"]; got != link {
		t.Fatalf("data-table-row-link = %#v; want %q", got, link)
	}
	if got := attrs["data-table-row-link-mode"]; got != "full" {
		t.Fatalf("data-table-row-link-mode = %#v; want full", got)
	}
	if _, ok := attrs["onclick"]; ok {
		t.Fatalf("full row link still emits executable onclick: %#v", attrs)
	}
	if _, ok := attrs["onauxclick"]; ok {
		t.Fatalf("full row link still emits executable onauxclick: %#v", attrs)
	}
}

func TestRowLinkAttrsKeepHTMXContractWithInertAuxiliaryTarget(t *testing.T) {
	link := "/people/o'connor"
	attrs := rowLinkAttrs(Row{Link: link})

	if got := attrs["data-table-row-link"]; got != link {
		t.Fatalf("data-table-row-link = %#v; want %q", got, link)
	}
	if got := attrs["hx-get"]; got != link {
		t.Fatalf("hx-get = %#v; want %q", got, link)
	}
	if _, ok := attrs["onauxclick"]; ok {
		t.Fatalf("HTMX row link still emits executable onauxclick: %#v", attrs)
	}
}
