package docspages

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentsIntegrationGuidePointsToAppShellsModule(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	require.NoError(t, agentsContent().Render(context.Background(), &output))
	html := output.String()

	assert.Contains(t, html, "Install the Skill")
	assert.Contains(t, html, `npx skills add araihu/goshtoso --skill using-goshtoso`)
	assert.Contains(t, html, "Use Without Installing")
	assert.Contains(t, html, `npx skills use araihu/goshtoso --skill using-goshtoso`)
	assert.Contains(t, html, `href="/modules/app-shells"`)
	assert.Contains(t, html, `Modules → App Shells`)
	assert.NotContains(t, html, "A Progressive Path")
	assert.NotContains(t, html, "Runtime Contracts the Browser Must Prove")
	assert.NotContains(t, html, "Build the invariant ledger")
	assert.NotContains(t, html, `depend on <a href="https://github.com/araihu/goshtoso-app-shells"`)
}
