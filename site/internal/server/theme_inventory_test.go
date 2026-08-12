package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	demothemes "github.com/araihu/goshtoso/site/internal/themes"
	"github.com/stretchr/testify/require"
)

func TestThemeInventoryRouteRendersEveryCatalogEntry(t *testing.T) {
	response := httptest.NewRecorder()
	(&Server{}).handleThemePage(response, httptest.NewRequest(http.MethodGet, "/docs/theme", nil))

	require.Equal(t, http.StatusOK, response.Code)
	body := response.Body.String()
	require.Equal(t, demothemes.Count(), strings.Count(body, `data-theme-key=`))
	require.Equal(t, 2, strings.Count(body, `data-theme-ownership="organization"`))
	require.Equal(t, demothemes.Count()-2, strings.Count(body, `data-theme-ownership="generic"`))
	for _, theme := range demothemes.All() {
		require.Equal(t, 1, strings.Count(body, `data-theme-key="`+theme.Key+`"`), "route inventory drift for %q", theme.Key)
	}
}
