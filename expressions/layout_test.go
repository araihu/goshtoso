package expressions_test

import (
	"github.com/araihu/goshtoso/expressions"
	"testing"
)

func TestLayoutMultipleMessageParameters(t *testing.T) {
	json := []byte(`{"Layout":{"MoveAnnouncement":"{position} / {containerLabel} / {itemLabel} {{literal}} {itemLabel}"}}`)
	yaml := []byte("Layout:\n  MoveAnnouncement: '{position} / {containerLabel} / {itemLabel} {{literal}} {itemLabel}'\n")
	for name, parse := range map[string]func() (expressions.Set, error){"json": func() (expressions.Set, error) { return expressions.ParseJSON(json) }, "yaml": func() (expressions.Set, error) { return expressions.ParseYAML(yaml) }} {
		t.Run(name, func(t *testing.T) {
			set, err := parse()
			if err != nil {
				t.Fatal(err)
			}
			got := set.Layout.MoveAnnouncement("A {position}", "B", "3")
			want := "3 / B / A {position} {literal} A {position}"
			if got != want {
				t.Fatalf("got %q", got)
			}
		})
	}
	for _, text := range []string{`{"Layout":{"MoveAnnouncement":"{unknown}"}}`, `{"Layout":{"MoveAnnouncement":"{position"}}`} {
		if _, err := expressions.ParseJSON([]byte(text)); err == nil {
			t.Fatal("accepted invalid placeholder")
		}
	}
}
