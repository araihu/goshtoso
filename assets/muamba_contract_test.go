package assets

import "testing"

func TestMuambaInventoryContainsRuntimeAndTailwindInputs(t *testing.T) {
	resources := MuambaResources()
	if len(resources) != 4 {
		t.Fatalf("resources = %d, want 4 embedded runtime resources", len(resources))
	}

	want := []string{"alpinejs", "htmx", "htmx-ext-sse", "htmx-ext-ws"}
	for _, name := range want {
		resource, ok := MuambaResourceByName(name)
		if !ok || resource.Version == "" || len(resource.Downloads) == 0 {
			t.Fatalf("resource %q = %#v, %t", name, resource, ok)
		}
		for _, download := range resource.Downloads {
			if download.Integrity == "" || download.Hash == "" {
				t.Fatalf("%s/%s is not locked: %#v", name, download.Name, download)
			}
		}
	}
	if _, ok := MuambaHash("alpinejs", "core-js"); !ok {
		t.Fatal("MuambaHash(alpinejs/core-js) missing")
	}
}
