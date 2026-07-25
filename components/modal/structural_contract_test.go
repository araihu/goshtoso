package modal

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

func TestAlertDialogUsesAlertDialogRole(t *testing.T) {
	html := renderStructuralModal(t, AlertDialog(AlertDialogConfig{
		ID:          "delete",
		Title:       "Delete?",
		ActionLabel: "Delete",
		Tone:        ToneDanger,
	}))

	require.Contains(t, html, `role="alertdialog"`)
	require.NotContains(t, html, "SecondaryLabel")
}

func renderStructuralModal(t *testing.T, component templ.Component) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, component.Render(context.Background(), &buf))
	return buf.String()
}
