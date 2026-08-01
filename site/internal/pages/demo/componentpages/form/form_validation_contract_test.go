package formpage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/form"
	formvalidation "github.com/araihu/goshtoso/components/form/validation"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/stretchr/testify/require"
)

func TestFormValidationOperationalLifecycle(t *testing.T) {
	newRequest := func() *http.Request {
		request := httptest.NewRequest(
			http.MethodPost,
			"/validate",
			strings.NewReader("email=person%40example.com"),
		)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return request
	}
	newDefinition := func() *formvalidation.FormDef {
		return &formvalidation.FormDef{
			FormID:   "profile",
			Endpoint: "/validate",
			Fields: map[string]*formvalidation.FieldDef{
				"email": {
					Name:       "email",
					FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}},
					OnChange:   true,
				},
			},
		}
	}

	var suppliedContext formvalidation.ValidationContext
	validResult := formvalidation.Handle(
		newRequest(),
		newDefinition(),
		func(ctx formvalidation.ValidationContext, _ string, _ *form.FieldGroupConfig) bool {
			suppliedContext = ctx
			return true
		},
	)
	require.NotNil(t, suppliedContext.FormValues)
	require.Equal(t, map[string]string{"email": "person@example.com"}, suppliedContext.FormValues)
	require.True(t, validResult.Valid, "Handle starts Result.Valid true when every hook succeeds")

	invalidResult := formvalidation.Handle(
		newRequest(),
		newDefinition(),
		func(formvalidation.ValidationContext, string, *form.FieldGroupConfig) bool {
			return false
		},
	)
	require.False(t, invalidResult.Valid, "one failed hook makes the complete Result invalid")
}

func TestFormValidationFieldGroupRequirementMatchesBindAndPopulateLifecycle(t *testing.T) {
	definition := &formvalidation.FormDef{
		FormID:   "profile",
		Endpoint: "/validate",
		Fields: map[string]*formvalidation.FieldDef{
			"email": {
				Name:       "email",
				FieldGroup: &form.FieldGroupConfig{Input: &textinput.Config{}},
				OnChange:   true,
			},
		},
	}
	definition.Bind()
	definition.PopulateValues(map[string]string{"email": "person@example.com"})
	require.NotNil(t, definition.Fields["email"].FieldGroup.Meta)
	require.Equal(t, "person@example.com", definition.Fields["email"].FieldGroup.Input.Value)

	invalid := &formvalidation.FormDef{
		Fields: map[string]*formvalidation.FieldDef{
			"email": {Name: "email"},
		},
	}
	require.Panics(t, invalid.Bind)
	require.Panics(t, func() {
		invalid.PopulateValues(map[string]string{"email": "person@example.com"})
	})
}

func TestFormValidationFieldResponseLifecycle(t *testing.T) {
	primary := &formvalidation.FieldDef{
		Name: "email",
		FieldGroup: &form.FieldGroupConfig{
			ID:    "email-field",
			Label: "Email",
			Input: &textinput.Config{ID: "email-input"},
		},
	}
	dependent := &formvalidation.FieldDef{
		Name: "confirm",
		FieldGroup: &form.FieldGroupConfig{
			ID:    "confirm-field",
			Label: "Confirm email",
			Input: &textinput.Config{ID: "confirm-input"},
			OOB:   true,
		},
	}
	recorder := httptest.NewRecorder()
	err := formvalidation.RenderFieldResponse(
		context.Background(),
		recorder,
		formvalidation.Result{Primary: primary, Dependents: []*formvalidation.FieldDef{dependent}},
	)
	require.NoError(t, err)
	require.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	require.Contains(t, recorder.Body.String(), `id="email-field"`)
	require.Contains(t, recorder.Body.String(), `id="confirm-field"`)
	require.Contains(t, recorder.Body.String(), `hx-swap-oob="true"`)
	require.False(t, dependent.FieldGroup.OOB, "RenderFieldResponse always resets dependent OOB false")

	renderError := errors.New("render failure")
	failing := &formvalidation.FieldDef{
		Name:       "failing",
		FieldGroup: &form.FieldGroupConfig{ID: "failing-field", OOB: true},
	}
	failingContext := templ.WithChildren(
		context.Background(),
		templ.ComponentFunc(func(context.Context, io.Writer) error {
			return renderError
		}),
	)
	err = formvalidation.RenderFieldResponse(
		failingContext,
		httptest.NewRecorder(),
		formvalidation.Result{Dependents: []*formvalidation.FieldDef{failing}},
	)
	require.ErrorIs(t, err, renderError)
	require.False(t, failing.FieldGroup.OOB, "error path always resets dependent OOB false")
}
