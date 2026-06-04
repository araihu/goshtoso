package badge

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderBadge(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Badge(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestBadge_LabelFieldRenders(t *testing.T) {
	html := renderBadge(t, Config{Label: "Live"})

	assert.Contains(t, html, "Live")
}
