package expressions_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/dropdown"
	"github.com/araihu/goshtoso/components/modal"
	"github.com/araihu/goshtoso/components/pagination"
	"github.com/araihu/goshtoso/components/schematree"
	"github.com/araihu/goshtoso/components/search"
	"github.com/araihu/goshtoso/expressions"
)

func TestRenderPrecedenceAndIsolation(t *testing.T) {
	renderer := expressions.NewRenderer(expressions.Set{Pagination: expressions.Pagination{PreviousAriaLabel: "system previous", NextAriaLabel: "system next"}})
	page := pagination.Pagination(pagination.Config{CurrentPage: 2, TotalPages: 3})
	ctx := expressions.With(context.Background(), expressions.Set{Pagination: expressions.Pagination{NextAriaLabel: "request next"}})
	component := page.WithExpressions(expressions.Pagination{PreviousAriaLabel: "component previous"})
	var out bytes.Buffer
	if err := renderer.Render(ctx, &out, component); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"component previous", "request next", `aria-label="pagination"`, `aria-label="page 3"`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in %s", want, out.String())
		}
	}
	out.Reset()
	if err := renderer.Render(context.Background(), &out, page); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "system previous") || strings.Contains(out.String(), "request next") || strings.Contains(out.String(), "component previous") {
		t.Fatal("overrides leaked across renders")
	}
}

func TestExistingConfigLabelsWin(t *testing.T) {
	ctx := expressions.With(context.Background(), expressions.Set{Search: expressions.Search{Label: "request", EmptyText: "request empty"}})
	component := search.Search(search.Config{Label: "explicit", EmptyText: "explicit empty"}).WithExpressions(expressions.Search{Label: "component", EmptyText: "component empty"})
	var out bytes.Buffer
	if err := component.Render(ctx, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "explicit") || strings.Contains(out.String(), ">component<") || strings.Contains(out.String(), ">request<") {
		t.Fatal(out.String())
	}
}

func TestConcurrentRenderingAndCancellation(t *testing.T) {
	renderer := expressions.NewRenderer(expressions.Set{})
	component := pagination.Pagination(pagination.Config{CurrentPage: 2, TotalPages: 3})
	var wg sync.WaitGroup
	for i := range 24 {
		wg.Go(func() {
			want := fmt.Sprintf("language-%d-page-3", i)
			ctx := expressions.With(context.Background(), expressions.Set{Pagination: expressions.Pagination{PageAriaLabel: func(n int) string { return fmt.Sprintf("language-%d-page-%d", i, n) }}})
			var out bytes.Buffer
			if err := renderer.Render(ctx, &out, component); err != nil {
				t.Error(err)
				return
			}
			if !strings.Contains(out.String(), want) {
				t.Errorf("missing request language %s", want)
			}
		})
	}
	wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sentinel := templ.ComponentFunc(func(ctx context.Context, _ io.Writer) error { return ctx.Err() })
	if err := renderer.Render(ctx, io.Discard, sentinel); err != context.Canceled {
		t.Fatalf("cancellation lost: %v", err)
	}
}

func TestExpressionsAreEscapedAndDefaultsAreValues(t *testing.T) {
	text := `Fechar "janela" <script>alert('x')</script>`
	var out bytes.Buffer
	if err := modal.Modal(modal.Config{}).WithExpressions(expressions.Modal{CloseAriaLabel: text}).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "<script>") || !strings.Contains(out.String(), "&lt;script&gt;") {
		t.Fatal("expression was not HTML escaped")
	}
	first := expressions.English()
	first.Modal.CloseAriaLabel = "changed"
	if expressions.English().Modal.CloseAriaLabel == "changed" {
		t.Fatal("English defaults mutated")
	}
}

func TestAdditionalRenderEntrypoints(t *testing.T) {
	cases := []struct {
		name      string
		component templ.Component
		want      string
	}{
		{"dropdown", dropdown.Dropdown(dropdown.Config{TriggerIconOnly: true, TriggerIcon: templ.Raw("<svg aria-hidden=\"true\"></svg>")}).WithExpressions(expressions.Dropdown{OpenMenuAriaLabel: "Abrir menu"}), "Abrir menu"},
		{"dot", badge.AnimatingDot(badge.ToneDanger).WithExpressions(expressions.Badge{NotificationAriaLabel: "Notificação"}), "Notificação"},
		{"dialog", modal.Dialog(modal.DialogConfig{ID: "dialog", Title: "Dialog"}).WithExpressions(expressions.Modal{DialogCloseAriaLabel: "Fechar diálogo"}), "Fechar diálogo"},
		{"schema tree", schematree.SchemaTree(schematree.Config{Nodes: []schematree.Node{{Name: "field", Required: true}}}).WithExpressions(expressions.SchemaTree{RequiredLabel: "Obrigatório"}), "Obrigatório"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := test.component.Render(context.Background(), &out); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), test.want) {
				t.Fatalf("missing %q in %s", test.want, out.String())
			}
		})
	}
}
