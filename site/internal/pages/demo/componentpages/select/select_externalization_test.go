package selectpage

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelectDraftRestoreUsesNamedSiteHelper(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	require.NoError(t, selectDraftRestorationPreview().Render(context.Background(), &output))
	require.Contains(t, output.String(), `x-on:click="window.goshtosoRestoreSelectDraft('draft-os','linux')"`)
	require.NotContains(t, output.String(), "input.dispatchEvent")
}
