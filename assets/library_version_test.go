package assets

import (
	"reflect"
	"runtime/debug"
	"testing"
)

func TestResolveGoshtosoVersionDistinguishesExactDevelopmentUnavailableAndReplaced(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want VersionInfo
	}{
		{
			name: "released dependency",
			info: &debug.BuildInfo{Deps: []*debug.Module{{
				Path: GoshtosoModulePath, Version: "v0.0.13", Sum: "h1:release",
			}}},
			ok: true,
			want: VersionInfo{
				ModulePath: GoshtosoModulePath,
				Status:     VersionExact,
				Version:    "v0.0.13",
				Sum:        "h1:release",
			},
		},
		{
			name: "exact main module",
			info: &debug.BuildInfo{Main: debug.Module{
				Path: GoshtosoModulePath, Version: "v0.0.14", Sum: "h1:main",
			}},
			ok: true,
			want: VersionInfo{
				ModulePath: GoshtosoModulePath,
				Status:     VersionExact,
				Version:    "v0.0.14",
				Sum:        "h1:main",
			},
		},
		{
			name: "development main module",
			info: &debug.BuildInfo{Main: debug.Module{
				Path: GoshtosoModulePath, Version: "(devel)",
			}},
			ok:   true,
			want: VersionInfo{ModulePath: GoshtosoModulePath, Status: VersionDevelopment},
		},
		{
			name: "empty development version",
			info: &debug.BuildInfo{Main: debug.Module{Path: GoshtosoModulePath}},
			ok:   true,
			want: VersionInfo{ModulePath: GoshtosoModulePath, Status: VersionDevelopment},
		},
		{
			name: "module absent",
			info: &debug.BuildInfo{Main: debug.Module{Path: "example.com/consumer", Version: "v1.0.0"}},
			ok:   true,
			want: VersionInfo{ModulePath: GoshtosoModulePath, Status: VersionUnavailable},
		},
		{
			name: "build info unavailable",
			info: nil,
			ok:   false,
			want: VersionInfo{ModulePath: GoshtosoModulePath, Status: VersionUnavailable},
		},
		{
			name: "local path replacement",
			info: &debug.BuildInfo{Deps: []*debug.Module{{
				Path: GoshtosoModulePath, Version: "v0.0.13", Sum: "h1:requested",
				Replace: &debug.Module{Path: "../goshtoso"},
			}}},
			ok: true,
			want: VersionInfo{
				ModulePath: GoshtosoModulePath,
				Status:     VersionReplaced,
				Requested: VersionReference{
					Path: GoshtosoModulePath, Version: "v0.0.13", Sum: "h1:requested",
				},
				Replacement: VersionReference{Path: "../goshtoso"},
			},
		},
		{
			name: "versioned replacement",
			info: &debug.BuildInfo{Deps: []*debug.Module{{
				Path: GoshtosoModulePath, Version: "v0.0.13", Sum: "h1:requested",
				Replace: &debug.Module{
					Path: "example.com/goshtoso-fork", Version: "v1.2.3", Sum: "h1:replacement",
				},
			}}},
			ok: true,
			want: VersionInfo{
				ModulePath: GoshtosoModulePath,
				Status:     VersionReplaced,
				Requested: VersionReference{
					Path: GoshtosoModulePath, Version: "v0.0.13", Sum: "h1:requested",
				},
				Replacement: VersionReference{
					Path: "example.com/goshtoso-fork", Version: "v1.2.3", Sum: "h1:replacement",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := resolveGoshtosoVersion(test.info, test.ok)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("resolveGoshtosoVersion() = %#v, want %#v", got, test.want)
			}
			if got.Status != VersionExact && (got.Version != "" || got.Sum != "") {
				t.Fatalf("non-exact status exposed release identity: %#v", got)
			}
		})
	}
}
