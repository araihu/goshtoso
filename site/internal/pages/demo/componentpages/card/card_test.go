package cardpage

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestCardDemoDocumentsPressedCustomMedia(t *testing.T) {
	var buffer bytes.Buffer
	if err := CardDemoPage().Render(context.Background(), &buffer); err != nil {
		t.Fatalf("CardDemoPage().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		`id="card-pressed-media"`,
		`data-card-media`,
		`hover:translate-y-1.5`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("card demo missing %q", want)
		}
	}
}
