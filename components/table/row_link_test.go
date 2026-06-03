package table

import (
	"strings"
	"testing"
)

func TestJSStringSingleEscapesRowLinks(t *testing.T) {
	got := jsStringSingle("/people/o'connor?x=1&next=\\home")
	want := `/people/o\'connor?x=1&next=\\home`
	if got != want {
		t.Fatalf("jsStringSingle = %q; want %q", got, want)
	}
}

func TestRowLinkAttrsEscapesFullNavigationURL(t *testing.T) {
	attrs := rowLinkAttrs(Row{
		Link:     "/people/o'connor?x=1&next=\\home",
		LinkMode: LinkFull,
	})

	onclick, _ := attrs["onclick"].(string)
	if strings.Contains(onclick, "o'connor") {
		t.Fatalf("onclick contains raw single quote: %q", onclick)
	}
	if !strings.Contains(onclick, `o\'connor`) {
		t.Fatalf("onclick missing escaped URL: %q", onclick)
	}
}

func TestRowLinkAttrsEscapesAuxClickURL(t *testing.T) {
	attrs := rowLinkAttrs(Row{Link: "/people/o'connor"})

	auxclick, _ := attrs["onauxclick"].(string)
	if strings.Contains(auxclick, "o'connor") {
		t.Fatalf("onauxclick contains raw single quote: %q", auxclick)
	}
	if !strings.Contains(auxclick, `o\'connor`) {
		t.Fatalf("onauxclick missing escaped URL: %q", auxclick)
	}
}
