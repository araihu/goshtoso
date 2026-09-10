package validation

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/araihu/goshtoso/components/form"
	"github.com/stretchr/testify/require"
)

type ownershipWriter struct {
	*httptest.ResponseRecorder
	check func()
	err   error
}

func (w ownershipWriter) Write(p []byte) (int, error) {
	w.check()
	if w.err != nil {
		return 0, w.err
	}
	return w.ResponseRecorder.Write(p)
}

func TestDependentResponseDoesNotMutateReusableConfiguration(t *testing.T) {
	for _, initial := range []bool{false, true} {
		for _, fail := range []bool{false, true} {
			first := &form.FieldGroupConfig{ID: "first", OOB: initial}
			second := &form.FieldGroupConfig{ID: "second", OOB: initial}
			result := Result{Dependents: []*FieldDef{{FieldGroup: first}, {FieldGroup: second}}}
			w := ownershipWriter{ResponseRecorder: httptest.NewRecorder(), check: func() {
				require.Equal(t, initial, first.OOB, "configuration must be immutable even during rendering")
				require.Equal(t, initial, second.OOB)
			}}
			if fail {
				w.err = errors.New("write failure")
			}
			err := RenderFieldResponse(t.Context(), w, result)
			if fail {
				require.ErrorIs(t, err, w.err)
			} else {
				require.NoError(t, err)
				require.Contains(t, w.Body.String(), `id="first"`)
				require.Contains(t, w.Body.String(), `id="second"`)
			}
			require.Equal(t, initial, first.OOB)
			require.Equal(t, initial, second.OOB)
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			require.Error(t, RenderFieldResponse(ctx, httptest.NewRecorder(), result))
			require.Equal(t, initial, first.OOB)
			require.NoError(t, RenderFieldResponse(t.Context(), httptest.NewRecorder(), result))
			require.Equal(t, initial, first.OOB)
		}
	}
}
