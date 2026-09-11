package expressions_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/schemaform"
	"github.com/araihu/goshtoso/components/toolbar"
	"github.com/araihu/goshtoso/expressions"
	"golang.org/x/net/html"
)

func TestSchemaLabelsPreserveRequiredAttributes(t *testing.T) {
	component := schemaform.Fields(schemaform.FieldsConfig{Fields: []schemaform.Field{
		{Path: "name", Name: "name", Label: "Name", Kind: schemaform.KindString, Required: true},
		{Path: "amount", Name: "amount", Label: "Amount", Kind: schemaform.KindNumber, Required: true},
		{Path: "choice", Name: "choice", Label: "Choice", Kind: schemaform.KindEnum, Enum: []string{"", "yes"}, Required: true},
	}}).WithExpressions(expressions.SchemaForm{RequiredAriaLabel: "Obrigatório"})
	var out bytes.Buffer
	if err := component.Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	doc, err := html.Parse(strings.NewReader(out.String()))
	if err != nil {
		t.Fatal(err)
	}
	required := 0
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		for _, attr := range node.Attr {
			if attr.Key == "required" {
				required++
			}
			if attr.Key == "Obrigatório" {
				t.Fatal("localized text became an HTML attribute")
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(doc)
	if required != 3 {
		t.Fatalf("want 3 required controls, got %d: %s", required, out.String())
	}
	if strings.Count(out.String(), `aria-label="Obrigatório"`) != 3 {
		t.Fatal("accessible labels not customized")
	}
}

func TestLegacyWhitespaceFallsThroughToRequest(t *testing.T) {
	ctx := expressions.With(context.Background(), expressions.Set{Toolbar: expressions.Toolbar{AriaLabel: "Ferramentas"}})
	var out bytes.Buffer
	if err := toolbar.Toolbar(toolbar.Config{Label: " \n "}).Render(ctx, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `aria-label="Ferramentas"`) {
		t.Fatal(out.String())
	}
}
