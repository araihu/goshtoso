package radio

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderRadio(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Radio(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestRadio_HelperTextRenders(t *testing.T) {
	html := renderRadio(t, Config{
		ID:         "daily",
		Name:       "cadence",
		Value:      "daily",
		Label:      "Daily",
		HelperText: "Runs every day",
	})

	assert.Contains(t, html, "Runs every day")
}
