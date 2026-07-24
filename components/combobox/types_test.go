package combobox

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	base := Config{
		ID:              "status",
		Name:            "status",
		ToggleEndpoint:  "/ui/combobox/status/toggle",
		OptionsEndpoint: "/ui/combobox/status/options",
		ClearEndpoint:   "/ui/combobox/status/clear",
		Source:          Source{LazyEndpoint: "/ui/combobox/status/options"},
	}

	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr string
	}{
		{name: "missing ID", mutate: func(c *Config) { c.ID = "" }, wantErr: "Config.ID"},
		{name: "missing Name", mutate: func(c *Config) { c.Name = "" }, wantErr: "Config.Name"},
		{name: "missing ToggleEndpoint", mutate: func(c *Config) { c.ToggleEndpoint = "" }, wantErr: "Config.ToggleEndpoint"},
		{name: "missing OptionsEndpoint", mutate: func(c *Config) { c.OptionsEndpoint = "" }, wantErr: "Config.OptionsEndpoint"},
		{name: "missing ClearEndpoint", mutate: func(c *Config) { c.ClearEndpoint = "" }, wantErr: "Config.ClearEndpoint"},
		{name: "no source", mutate: func(c *Config) { c.Source = Source{} }, wantErr: "Source"},
		{name: "both sources", mutate: func(c *Config) { c.Source.Static = []Option{{Value: "a"}} }, wantErr: "both Static and LazyEndpoint"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := base
			tc.mutate(&c)
			err := c.Validate()
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestState_IsSelected(t *testing.T) {
	s := State{Selected: []string{"a", "b"}}
	assert.True(t, s.isSelected("a"))
	assert.True(t, s.isSelected("b"))
	assert.False(t, s.isSelected("c"))
	assert.False(t, s.isSelected(""))
}

func TestConfig_DepsSelector(t *testing.T) {
	assert.Equal(t, "", Config{}.depsSelector())
	assert.Equal(t, "[name='provider']", Config{DependsOn: []string{"provider"}}.depsSelector())
	assert.Equal(t, "[name='provider'],[name='zone']", Config{DependsOn: []string{"provider", "zone"}}.depsSelector())
}

func TestConfig_HXIncludeSelector(t *testing.T) {
	base := "closest [data-combobox] input[type=hidden]"
	assert.Equal(t, base, Config{}.hxIncludeSelector())
	assert.Equal(t, base+",[name='provider']", Config{DependsOn: []string{"provider"}}.hxIncludeSelector())
}

func TestConfig_Validate_ClientMode(t *testing.T) {
	base := Config{
		ID:     "status",
		Name:   "status",
		Source: Source{Static: []Option{{Value: "a", Label: "A"}}},
	}

	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr string
	}{
		{name: "client mode valid without endpoints", mutate: func(c *Config) {}, wantErr: ""},
		{name: "client mode ignores server endpoints", mutate: func(c *Config) {
			c.ToggleEndpoint = "/t"
			c.OptionsEndpoint = "/o"
			c.ClearEndpoint = "/c"
		}, wantErr: ""},
		{name: "lazy mode still requires endpoints", mutate: func(c *Config) {
			c.Source = Source{LazyEndpoint: "/x"}
		}, wantErr: "Config.ToggleEndpoint"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := base
			tc.mutate(&c)
			err := c.Validate()
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestConfig_IsClientMode(t *testing.T) {
	static := Config{Source: Source{Static: []Option{{Value: "a"}}}}
	assert.True(t, static.isClientMode())

	lazy := Config{Source: Source{LazyEndpoint: "/x"}}
	assert.False(t, lazy.isClientMode())
}
