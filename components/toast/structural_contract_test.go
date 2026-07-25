package toast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageToastHasItsOwnConfig(t *testing.T) {
	html := renderToastComponent(t, MessageToast(MessageConfig{
		Sender:      Sender{Name: "Ada"},
		Message:     "Review ready",
		ActionLabel: "Open",
	}))

	require.Contains(t, html, "Ada")
	require.Contains(t, html, "Open")
	require.NotContains(t, html, ">Reply<")
}
