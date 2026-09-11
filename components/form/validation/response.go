package validation

import (
	"context"
	"net/http"

	"github.com/araihu/goshtoso/components/form"
)

// RenderFieldResponse writes the HTMX response for field-level validation.
// Renders the primary field as the main swap, and dependents as OOB swaps.
func RenderFieldResponse(ctx context.Context, w http.ResponseWriter, result Result) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render primary field (outerHTML swap target)
	if result.Primary != nil {
		if err := form.FieldGroup(*result.Primary.FieldGroup).Render(ctx, w); err != nil {
			return err
		}
	}

	// Render dependent fields with OOB swap
	for _, dep := range result.Dependents {
		// Routing belongs to this response, not the reusable field definition.
		field := *dep.FieldGroup
		field.OOB = true
		if err := form.FieldGroup(field).Render(ctx, w); err != nil {
			return err
		}
	}

	return nil
}
