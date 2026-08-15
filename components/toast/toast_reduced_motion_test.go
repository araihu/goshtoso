package toast

import (
	"html"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToastRenderedTransitionsProvideReducedMotionContract(t *testing.T) {
	containerHTML := html.UnescapeString(renderToastComponent(t, ToastContainer(ContainerConfig{})))
	serverHTML := html.UnescapeString(renderToast(t, Config{Tone: ToneSuccess}))
	messageHTML := html.UnescapeString(renderToastComponent(t, MessageToast(MessageConfig{Message: "Review ready"})))

	allHTML := containerHTML + serverHTML + messageHTML
	enter := `x-transition:enter="transition duration-300 ease-out motion-reduce:transition-none"`
	leave := `x-transition:leave="transition duration-300 ease-in motion-reduce:transition-none"`
	enterStart := `x-transition:enter-start="translate-y-8 motion-reduce:translate-y-0"`
	leaveEnd := `x-transition:leave-end="-translate-x-24 opacity-0 md:translate-x-24 motion-reduce:translate-x-0 md:motion-reduce:translate-x-0 motion-reduce:opacity-100"`

	const variantCount = 7 // five client templates plus one server template per kind
	require.Equal(t, variantCount, strings.Count(allHTML, enter), "every client/server toast variant needs reduced-motion enter")
	require.Equal(t, variantCount, strings.Count(allHTML, leave), "every client/server toast variant needs reduced-motion leave")
	require.Equal(t, variantCount, strings.Count(allHTML, enterStart), "every client/server toast variant needs an immediate enter state")
	require.Equal(t, variantCount, strings.Count(allHTML, leaveEnd), "every client/server toast variant needs an immediate leave state")
}
