//go:build e2e && full

package e2e

import (
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStopServerUsesShutdownEndpointWhenConfigured(t *testing.T) {
	var requested bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = true
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/__e2e/shutdown", r.URL.Path)
		require.Equal(t, "secret", r.URL.Query().Get("token"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := stopServer(&exec.Cmd{}, srv.URL, "secret")

	require.NoError(t, err)
	require.True(t, requested)
}

func TestE2ECoverPkgIncludesServerMainPackage(t *testing.T) {
	got := e2eCoverPkg("github.com/araihu/goshtoso/components/button,github.com/araihu/goshtoso/components/table")

	require.Equal(t,
		"github.com/araihu/goshtoso/site/cmd/server,github.com/araihu/goshtoso/components/button,github.com/araihu/goshtoso/components/table",
		got,
	)
}
