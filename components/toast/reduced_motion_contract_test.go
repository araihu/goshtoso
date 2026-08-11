package toast

import (
	"html"
	"strings"
	"testing"
)

func TestToastRendersReducedMotionTransitionsAndRemovalDelay(t *testing.T) {
	containerData := containerAlpineData(ContainerConfig{})
	for _, want := range []string{
		"prefers-reduced-motion: reduce",
		"addEventListener('change'",
		"removeEventListener('change'",
		"reducedMotion",
		"? 0 : 400",
	} {
		if !strings.Contains(containerData, want) {
			t.Errorf("Toast container runtime missing reduced-motion contract %q", want)
		}
	}

	containerHTML := html.UnescapeString(renderToastComponent(t, ToastContainer(ContainerConfig{})))
	if count := strings.Count(containerHTML, "motion-reduce:transition-none"); count < 2 {
		t.Errorf("client Toast transitions have %d reduced-motion overrides; want at least 2", count)
	}

	messageHTML := html.UnescapeString(renderToastComponent(t, MessageToast(MessageConfig{
		Message:         "Review ready",
		DisplayDuration: -1,
	})))
	for _, want := range []string{"motion-reduce:transition-none", "reducedMotion", "? 0 : 400"} {
		if !strings.Contains(messageHTML, want) {
			t.Errorf("server message Toast missing reduced-motion contract %q", want)
		}
	}
}
