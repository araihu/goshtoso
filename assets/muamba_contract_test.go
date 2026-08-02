package assets

import "testing"

func TestMuambaInventoryContainsRuntimeAndTailwindInputs(t *testing.T) {
	resources := MuambaResources()
	if len(resources) != 5 {
		t.Fatalf("resources = %d, want 4 runtime resources plus Tailwind license metadata", len(resources))
	}

	want := []string{"alpinejs", "htmx", "htmx-ext-sse", "htmx-ext-ws", "tailwindcss"}
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
	if resource, ok := MuambaResourceByName("tailwindcss"); !ok || resource.Version != "4.3.3" {
		t.Fatalf("tailwindcss resource = %#v, %t", resource, ok)
	}
	if _, ok := MuambaHash("tailwindcss", "license"); !ok {
		t.Fatal("MuambaHash(tailwindcss/license) missing")
	}
	if _, ok := MuambaHash("tailwindcss", "cli"); ok {
		t.Fatal("Tailwind executable must not be embedded in package assets")
	}
}

func TestRuntimeHashMatchesMuambaInventory(t *testing.T) {
	tests := []struct {
		role     RuntimeAssetRole
		resource string
		download string
	}{
		{RuntimeRoleAlpineCollapse, "alpinejs", "collapse-js"},
		{RuntimeRoleAlpineFocus, "alpinejs", "focus-js"},
		{RuntimeRoleAlpineMask, "alpinejs", "mask-js"},
		{RuntimeRoleAlpineJS, "alpinejs", "core-js"},
		{RuntimeRoleHTMX, "htmx", "core-js"},
		{RuntimeRoleHTMXExtSSE, "htmx-ext-sse", "runtime-js"},
		{RuntimeRoleHTMXExtWS, "htmx-ext-ws", "runtime-js"},
	}
	for _, test := range tests {
		got, ok := RuntimeHash(test.role)
		want, wantOK := MuambaHash(test.resource, test.download)
		if !ok || !wantOK || got != want {
			t.Errorf("RuntimeHash(%q) = %q, %t; MuambaHash(%s/%s) = %q, %t", test.role, got, ok, test.resource, test.download, want, wantOK)
		}
	}
	for _, role := range []RuntimeAssetRole{RuntimeRoleFirstParty, RuntimeRoleDependencyLoader, "unknown"} {
		if hash, ok := RuntimeHash(role); ok || hash != "" {
			t.Errorf("RuntimeHash(%q) = %q, %t; want missing", role, hash, ok)
		}
	}
}
