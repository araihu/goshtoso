package combobox

import (
	"os/exec"
	"testing"
)

func TestControlledClientOptionsAndSearch(t *testing.T) {
	node, e := exec.LookPath("node")
	if e != nil {
		t.Skip("Node unavailable")
	}
	cmd := exec.CommandContext(t.Context(), node, "--test", "client_runtime_test.mjs")
	if output, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("controlled combobox: %v\n%s", e, output)
	}
}
