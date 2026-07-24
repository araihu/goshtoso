package accordion

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestCoverageRenderConfiguredAccordion(t *testing.T) {
	var buf bytes.Buffer
	err := Accordion(AccordionConfig{
		ID:            "faq",
		RootClass:     "custom-root",
		Appearance:    AppearanceSplit,
		AllowMultiple: true,
		Items: []AccordionItem{
			{
				ID:                "billing",
				Title:             "Billing",
				Content:           templ.Raw("Billing details"),
				Icon:              templ.Raw(`<span data-testid="billing-icon"></span>`),
				InitiallyExpanded: true,
			},
			{
				Title:    "Shipping",
				Content:  templ.Raw("Shipping details"),
				Disabled: true,
			},
			{
				ID:                "returns",
				Title:             "Returns",
				Content:           templ.Raw("Returns details"),
				InitiallyExpanded: true,
			},
		},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render configured accordion: %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="faq"`,
		`custom-root`,
		`x-data="{ opened: [true,false,true], allowMultiple: true`,
		`id="controls-billing"`,
		`aria-controls="content-billing"`,
		`:aria-expanded="isOpen(0) ? &#39;true&#39; : &#39;false&#39;"`,
		`x-show="isOpen(0)"`,
		`role="region"`,
		`aria-labelledby="controls-billing"`,
		`data-testid="billing-icon"`,
		`id="controls-accordion-item-1"`,
		`aria-controls="content-accordion-item-1"`,
		`disabled`,
		`Billing details`,
		`Shipping details`,
		`Returns details`,
		`rounded-radius border border-outline bg-surface-alt/40`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("configured render missing %q in:\n%s", want, html)
		}
	}
}

func TestCoverageRenderDefaultAccordion(t *testing.T) {
	var buf bytes.Buffer
	err := Accordion(AccordionConfig{
		Items: []AccordionItem{
			{Title: "First", Content: templ.Raw("First content")},
			{Title: "Second", Content: templ.Raw("Second content")},
		},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render default accordion: %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="accordion"`,
		`x-data="{ opened: [false,false], allowMultiple: false`,
		`bg-surface-alt/40`,
		`id="controls-accordion-item-0"`,
		`id="content-accordion-item-1"`,
		`@click="toggle(1)"`,
		`:class="isOpen(1) ? &#39;text-on-surface-strong`,
		`x-collapse`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("default render missing %q in:\n%s", want, html)
		}
	}
}

func TestCoverageAccordionClassHelpers(t *testing.T) {
	tests := []struct {
		name string
		cfg  AccordionConfig
		want string
	}{
		{name: "default", cfg: AccordionConfig{}, want: "bg-surface-alt/40"},
		{name: "plain", cfg: AccordionConfig{Appearance: AppearancePlain}, want: "bg-surface dark:bg-surface-dark"},
		{name: "split", cfg: AccordionConfig{Appearance: AppearanceSplit}, want: "flex w-full flex-col gap-4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ContainerClasses(); !strings.Contains(got, tt.want) {
				t.Fatalf("ContainerClasses() = %q, want it to contain %q", got, tt.want)
			}
		})
	}

	defaultData := AccordionItemData{}
	if got := defaultData.ItemContainerClasses(); got != "" {
		t.Fatalf("default ItemContainerClasses() = %q, want empty", got)
	}
	if got := defaultData.ItemButtonClasses(); !strings.Contains(got, "bg-surface-alt hover:bg-surface-alt/75") {
		t.Fatalf("default ItemButtonClasses() = %q, want default background classes", got)
	}

	plainData := AccordionItemData{Appearance: AppearancePlain}
	if got := plainData.ItemButtonClasses(); !strings.Contains(got, "bg-surface hover:bg-surface-alt") {
		t.Fatalf("plain ItemButtonClasses() = %q, want plain surface classes", got)
	}

	splitData := AccordionItemData{Appearance: AppearanceSplit}
	if got := splitData.ItemContainerClasses(); !strings.Contains(got, "rounded-radius border border-outline") {
		t.Fatalf("split ItemContainerClasses() = %q, want split card classes", got)
	}
	if got := splitData.ExpandedClasses(); !strings.Contains(got, "font-bold") {
		t.Fatalf("ExpandedClasses() = %q, want bold text", got)
	}
	if got := splitData.CollapsedClasses(); !strings.Contains(got, "font-medium") {
		t.Fatalf("CollapsedClasses() = %q, want medium text", got)
	}
	if got := splitData.ContentClasses(); !strings.Contains(got, "text-pretty") {
		t.Fatalf("ContentClasses() = %q, want content typography", got)
	}
}

func TestZeroValueAllowsOnlyOneOpen(t *testing.T) {
	if got := generateAlpineData(AccordionConfig{}); !strings.Contains(got, "allowMultiple: false") {
		t.Fatalf("zero-value accordion must allow only one open item; got %q", got)
	}
}

func TestCoverageGenerateAlpineData(t *testing.T) {
	got := generateAlpineData(AccordionConfig{
		AllowMultiple: true,
		Items: []AccordionItem{
			{InitiallyExpanded: true},
			{},
			{InitiallyExpanded: true},
		},
	})

	for _, want := range []string{
		`opened: [true,false,true]`,
		`allowMultiple: true`,
		`this.opened[index] = !this.opened[index]`,
		`this.opened = this.opened.map(() => false)`,
		`isOpen(index) { return this.opened[index]; }`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generateAlpineData() missing %q in %q", want, got)
		}
	}
}
