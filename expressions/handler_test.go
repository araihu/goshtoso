package expressions_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/combobox"
	"github.com/araihu/goshtoso/expressions"
)

func TestComboboxHandlerUsesRequestExpressions(t *testing.T) {
	cfg := combobox.Config{ID: fmt.Sprintf("expressions-%p", t), Name: "choice", Source: combobox.Source{LazyEndpoint: "/options"}, OptionsEndpoint: "/options", ToggleEndpoint: "/toggle", ClearEndpoint: "/clear"}
	handler := combobox.Handler(cfg, func(_ context.Context, query string, _ map[string]string) ([]combobox.Option, error) {
		if query == "fail" {
			return nil, errors.New("provider diagnostic")
		}
		return nil, nil
	})
	for _, test := range []struct {
		query, want string
		status      int
	}{{"", "Sem resultados", http.StatusOK}, {"fail", "Tentar novamente", http.StatusBadGateway}} {
		req := httptest.NewRequest(http.MethodGet, "/options?q="+test.query, nil)
		ctx := expressions.With(req.Context(), expressions.Set{Combobox: expressions.Combobox{EmptyText: "Sem resultados", ErrorText: "Falha ao carregar", RetryLabel: "Tentar novamente"}})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req.WithContext(ctx))
		if rec.Code != test.status || !strings.Contains(rec.Body.String(), test.want) {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	}
}
