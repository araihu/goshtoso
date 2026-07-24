package pagination

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoveragePageWindowShapes(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want []PageItem
	}{
		{
			name: "no pages",
			cfg:  Config{CurrentPage: 1, TotalPages: 0},
		},
		{
			name: "small count shows every page",
			cfg:  Config{CurrentPage: 3, TotalPages: 5},
			want: []PageItem{
				{Page: 1}, {Page: 2}, {Page: 3, IsCurrent: true}, {Page: 4}, {Page: 5},
			},
		},
		{
			name: "near beginning has trailing ellipsis",
			cfg:  Config{CurrentPage: 2, TotalPages: 30},
			want: []PageItem{
				{Page: 1}, {Page: 2, IsCurrent: true}, {Page: 3}, {Page: 4}, {IsEllipsis: true}, {Page: 30},
			},
		},
		{
			name: "middle has both ellipses",
			cfg:  Config{CurrentPage: 15, TotalPages: 30},
			want: []PageItem{
				{Page: 1}, {IsEllipsis: true}, {Page: 14}, {Page: 15, IsCurrent: true}, {Page: 16}, {IsEllipsis: true}, {Page: 30},
			},
		},
		{
			name: "near end has leading ellipsis",
			cfg:  Config{CurrentPage: 29, TotalPages: 30},
			want: []PageItem{
				{Page: 1}, {IsEllipsis: true}, {Page: 27}, {Page: 28}, {Page: 29, IsCurrent: true}, {Page: 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cfg.Pages())
		})
	}
}

func TestCoveragePreviousNextAndSwapDefaults(t *testing.T) {
	first := Config{CurrentPage: 1, TotalPages: 3}
	assert.Equal(t, 1, first.PreviousPage())
	assert.Equal(t, 2, first.NextPage())

	last := Config{CurrentPage: 3, TotalPages: 3}
	assert.Equal(t, 2, last.PreviousPage())
	assert.Equal(t, 3, last.NextPage())

	assert.Equal(t, "innerHTML", Config{}.SwapStrategy())
	assert.Equal(t, "outerHTML", Config{HTMX: &HTMXConfig{Swap: "outerHTML"}}.SwapStrategy())
}

func TestCoverageRenderEllipsisAndHTMXBranches(t *testing.T) {
	rendered := renderPagination(t, Config{
		ID:          "paged-items",
		Mode:        ModeEllipsis,
		CurrentPage: 15,
		TotalPages:  30,
		BaseURL:     "/items?filter=active",
		NavClass:    "justify-center",
		HTMX:        &HTMXConfig{Target: "#items", Swap: "outerHTML"},
	})

	assert.Contains(t, rendered, `id="paged-items"`)
	assert.Contains(t, rendered, `aria-label="pagination"`)
	assert.Contains(t, rendered, `justify-center`)
	assert.Contains(t, rendered, `href="/items?filter=active&amp;page=14"`)
	assert.Contains(t, rendered, `hx-get="/items?filter=active&amp;page=14"`)
	assert.Contains(t, rendered, `hx-target="#items"`)
	assert.Contains(t, rendered, `hx-swap="outerHTML"`)
	assert.Contains(t, rendered, `aria-current="page"`)
	assert.Contains(t, rendered, `aria-label="more pages"`)
	assert.Contains(t, rendered, `stroke-linecap="round"`)
	assert.Equal(t, 2, strings.Count(rendered, `aria-label="more pages"`))
}

func TestCoverageRenderSimpleLastPageDisablesNext(t *testing.T) {
	rendered := renderPagination(t, Config{
		Mode:        ModeSimple,
		CurrentPage: 4,
		TotalPages:  4,
		BaseURL:     "/items",
	})

	assert.Contains(t, rendered, `href="/items?page=3"`)
	assert.Contains(t, rendered, `aria-label="previous page"`)
	assert.Contains(t, rendered, `aria-disabled="true"`)
	assert.Contains(t, rendered, `cursor-not-allowed`)
	assert.NotContains(t, rendered, `aria-label="page 1"`)
}
