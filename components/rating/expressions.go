package rating

import (
	"context"
	"github.com/araihu/goshtoso/expressions"
)

func expressionValueLabel(ctx context.Context, appearance Appearance, value int) string {
	text := expressions.From(ctx).Rating
	if appearance == AppearanceEmoji {
		labels := []string{text.VeryDissatisfiedLabel, text.DissatisfiedLabel, text.NeutralLabel, text.SatisfiedLabel, text.VerySatisfiedLabel}
		if value >= 1 && value <= len(labels) {
			return labels[value-1]
		}
	}
	return text.StarLabel(value)
}
