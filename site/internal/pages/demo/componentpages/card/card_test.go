package cardpage

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPressedMediaPreviewUsesReleasedCardComposition(t *testing.T) {
	t.Parallel()
	var buffer bytes.Buffer
	if err := cardPressedMediaPreview().Render(context.Background(), &buffer); err != nil {
		t.Fatalf("render preview: %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		`id="card-pressed-media"`,
		`data-card-media`,
		`hover:translate-y-1.5`,
		`motion-reduce:transition-none`,
		`Atlas`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("pressed media preview missing %q", want)
		}
	}
}
