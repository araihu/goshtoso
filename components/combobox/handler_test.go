package combobox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func staticProvider(opts []Option) OptionsProvider {
	return func(ctx context.Context, search string, deps map[string]string) ([]Option, error) {
		if search == "" {
			return opts, nil
		}
		out := []Option{}
		for _, o := range opts {
			if strings.Contains(strings.ToLower(o.Label), strings.ToLower(search)) {
				out = append(out, o)
			}
		}
		return out, nil
	}
}

func TestHandler_GetOptions_RendersListForCurrentSearchAndSelection(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "users", Name: "user", Mode: ModeMultiple,
		ToggleEndpoint: "/c/toggle", OptionsEndpoint: "/c/options", ClearEndpoint: "/c/clear",
		Source:       Source{LazyEndpoint: "/c/options"},
		EnableSearch: true,
	}
	provider := staticProvider([]Option{
		{Value: "alice", Label: "Alice"},
		{Value: "bob", Label: "Bob"},
		{Value: "albert", Label: "Albert"},
	})
	h := Handler(cfg, provider)

	form := url.Values{}
	form.Set("q", "al")
	form.Add("user", "bob") // currently selected, filtered out by search

	req := httptest.NewRequest(http.MethodGet, "/c/options?"+form.Encode(), nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	html := string(body)

	// Filter matched Alice and Albert.
	assert.Contains(t, html, `data-value="alice"`)
	assert.Contains(t, html, `data-value="albert"`)
	assert.NotContains(t, html, `data-value="bob"`)
	// Aria-selected reflects the carried 'bob' selection — but bob not rendered. Check no alice/albert checked.
	assert.Contains(t, html, `data-value="alice" aria-selected="false"`)
}

func TestHandler_PostToggle_Multi_AppendsValue(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		ToggleEndpoint: "/status/toggle", OptionsEndpoint: "/status/options", ClearEndpoint: "/status/clear",
		Source: Source{Static: []Option{{Value: "creating"}, {Value: "running"}, {Value: "failed"}}},
	}
	h := Handler(cfg, staticProvider(cfg.Source.Static))

	form := url.Values{}
	form.Add("status", "creating")
	form.Set("value", "running")

	req := httptest.NewRequest(http.MethodPost, "/status/toggle", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	// Body now has two hidden inputs.
	body := rec.Body.String()
	assert.Equal(t, 2, strings.Count(body, `type="hidden"`))
	assert.Contains(t, body, `value="creating"`)
	assert.Contains(t, body, `value="running"`)

	// HX-Trigger header carries new values.
	trigger := rec.Header().Get("HX-Trigger")
	require.NotEmpty(t, trigger)
	assert.Contains(t, trigger, `"id":"status"`)
	assert.Contains(t, trigger, `"creating"`)
	assert.Contains(t, trigger, `"running"`)
}

func TestHandler_PostToggle_Multi_RemovesAlreadySelected(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		ToggleEndpoint: "/status/toggle", OptionsEndpoint: "/status/options", ClearEndpoint: "/status/clear",
		Source: Source{Static: []Option{{Value: "running"}, {Value: "creating"}}},
	}
	h := Handler(cfg, staticProvider(cfg.Source.Static))

	form := url.Values{}
	form.Add("status", "creating")
	form.Add("status", "running")
	form.Set("value", "running")

	req := httptest.NewRequest(http.MethodPost, "/status/toggle", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.Equal(t, 1, strings.Count(body, `type="hidden"`))
	assert.Contains(t, body, `type="hidden" name="status" value="creating"`)
	assert.NotContains(t, body, `type="hidden" name="status" value="running"`)
}

func TestHandler_PostToggle_Single_ReplacesValue(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "provider", Name: "provider", Mode: ModeSingle,
		ToggleEndpoint: "/provider/toggle", OptionsEndpoint: "/provider/options", ClearEndpoint: "/provider/clear",
		Source: Source{Static: []Option{{Value: "maas"}, {Value: "eks"}}},
	}
	h := Handler(cfg, staticProvider(cfg.Source.Static))

	form := url.Values{}
	form.Add("provider", "maas")
	form.Set("value", "eks")

	req := httptest.NewRequest(http.MethodPost, "/provider/toggle", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.Equal(t, 1, strings.Count(body, `type="hidden"`))
	assert.Contains(t, body, `type="hidden" name="provider" value="eks"`)
	assert.NotContains(t, body, `type="hidden" name="provider" value="maas"`)

	trigger := rec.Header().Get("HX-Trigger")
	assert.Contains(t, trigger, `"eks"`)
	assert.NotContains(t, trigger, `"maas"`)
}

func TestHandler_PostClear_EmptiesSelection(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "status", Name: "status", Mode: ModeMultiple,
		ToggleEndpoint: "/status/toggle", OptionsEndpoint: "/status/options", ClearEndpoint: "/status/clear",
		Source: Source{Static: []Option{{Value: "a"}, {Value: "b"}}},
	}
	h := Handler(cfg, staticProvider(cfg.Source.Static))

	form := url.Values{}
	form.Add("status", "a")
	form.Add("status", "b")

	req := httptest.NewRequest(http.MethodPost, "/status/clear", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.NotContains(t, body, `type="hidden"`)

	trigger := rec.Header().Get("HX-Trigger")
	assert.Contains(t, trigger, `"values":null`)
}

func TestHandler_ProviderError_Returns502WithRetarget(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "users", Name: "user", Mode: ModeMultiple,
		ToggleEndpoint: "/users/toggle", OptionsEndpoint: "/users/options", ClearEndpoint: "/users/clear",
		Source: Source{LazyEndpoint: "/users/options"},
	}
	provider := OptionsProvider(func(ctx context.Context, search string, deps map[string]string) ([]Option, error) {
		return nil, assert.AnError
	})
	h := Handler(cfg, provider)

	req := httptest.NewRequest(http.MethodGet, "/users/options", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Equal(t, "#users-options", rec.Header().Get("HX-Retarget"))
	assert.Contains(t, rec.Body.String(), `Failed to load`)
}

func TestHandler_ProviderError_Escapes(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "x<script>", Name: "x",
		ToggleEndpoint: "/t/toggle", OptionsEndpoint: "/t/options", ClearEndpoint: "/t/clear",
		Source: Source{LazyEndpoint: "/t/options"},
	}
	provider := func(_ context.Context, _ string, _ map[string]string) ([]Option, error) {
		return nil, fmt.Errorf("boom")
	}
	h := Handler(cfg, provider)

	req := httptest.NewRequest(http.MethodGet, "/t/options", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.NotContains(t, body, `<script>`, "cfg.ID must be escaped")
	assert.Contains(t, body, `&lt;script&gt;`)
	assert.Equal(t, "#x<script>-options", rec.Header().Get("HX-Retarget"), "HX-Retarget is a header, not HTML — not escaped")
}

func TestHandler_CascadeInvalidation_DropsStaleSelection(t *testing.T) {
	resetRegistry()
	cfg := Config{
		ID: "region", Name: "region", Mode: ModeMultiple,
		ToggleEndpoint: "/region/toggle", OptionsEndpoint: "/region/options", ClearEndpoint: "/region/clear",
		Source:    Source{LazyEndpoint: "/region/options"},
		DependsOn: []string{"provider"},
	}
	provider := OptionsProvider(func(ctx context.Context, search string, deps map[string]string) ([]Option, error) {
		if deps["provider"] == "maas" {
			return []Option{{Value: "us-east", Label: "US East"}}, nil
		}
		return []Option{{Value: "eu-central-1", Label: "EU Central 1"}}, nil
	})
	h := Handler(cfg, provider)

	// Client had region=eu-central-1 selected under provider=eks; provider now maas.
	form := url.Values{}
	form.Add("region", "eu-central-1")
	form.Set("provider", "maas")
	form.Set("value", "us-east")

	req := httptest.NewRequest(http.MethodPost, "/region/toggle", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.NotContains(t, body, `type="hidden" name="region" value="eu-central-1"`, "stale region must be dropped")
	assert.Contains(t, body, `type="hidden" name="region" value="us-east"`)
}
