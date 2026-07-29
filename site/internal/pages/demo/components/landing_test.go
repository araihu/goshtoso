package components

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"

	siteassets "github.com/araihu/goshtoso/site/assets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLandingPageRendersRegistryComponentCount(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, LandingPage().Render(context.Background(), &buf))

	expected := strconv.Itoa(componentCount()) + " components"
	assert.Equal(t, 2, strings.Count(buf.String(), expected))
}

func TestLandingPageUsesSiteOwnedBootstrapAndProviders(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, LandingPage().Render(context.Background(), &buf))

	html := buf.String()
	assert.Contains(t, html, `data-demo-storage-policy="strict"`)
	assert.Contains(t, html, `data-demo-theme-bootstrap`)
	assert.Contains(t, html, `<script src="`+siteassets.DemoBundleURL+`"></script>`)
	assert.Contains(t, html, `x-data="demoStorageConsent"`)
	assert.NotContains(t, html, "window.goshtosoStorageConsent={")
	assert.NotContains(t, html, "localStorage.getItem('darkMode')")
}
