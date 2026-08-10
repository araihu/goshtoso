package themes_test

import (
	"testing"

	"github.com/araihu/goshtoso/themes"
)

func TestPublicConsumerContractCompiles(t *testing.T) {
	keys := [...]themes.Key{
		themes.Key90s,
		themes.KeyAraiHu,
		themes.KeyArctic,
		themes.KeyChristmas,
		themes.KeyDracula,
		themes.KeyGoshtoso,
		themes.KeyHalloween,
		themes.KeyHighContrast,
		themes.KeyIndustrial,
		themes.KeyMinimal,
		themes.KeyModern,
		themes.KeyNeoBrutalism,
		themes.KeyNews,
		themes.KeyPastel,
		themes.KeyPrototype,
		themes.KeyZombie,
	}

	catalog := themes.BuiltIn()
	if len(catalog) != len(keys) {
		t.Fatalf("len(themes.BuiltIn()) = %d, want %d public keys", len(catalog), len(keys))
	}
	for index, key := range keys {
		if catalog[index].Key != key {
			t.Fatalf("themes.BuiltIn()[%d].Key = %q, want %q", index, catalog[index].Key, key)
		}
		if catalog[index].Label == "" {
			t.Fatalf("themes.BuiltIn()[%d].Label is empty", index)
		}
	}
}
