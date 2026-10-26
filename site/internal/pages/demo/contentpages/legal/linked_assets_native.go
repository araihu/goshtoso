//go:build !js

package legalpages

import (
	"github.com/araihu/goshtoso/assets"
	"net/http"
	"net/http/httptest"
)

func linkedAssetServes(assetURL string) bool {
	request := httptest.NewRequest(http.MethodGet, assetURL, nil)
	response := httptest.NewRecorder()
	assets.Handler().ServeHTTP(response, request)
	return response.Code == http.StatusOK
}
