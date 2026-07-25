package toast

import (
	"bytes"
	"context"
	"html"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderToast(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Toast(cfg).Render(context.Background(), &buf))
	return buf.String()
}

func TestToast_UniqueIDIsConcurrencySafe(t *testing.T) {
	idCounter.Store(0)
	const count = 64
	ids := make(chan int64, count)
	var wg sync.WaitGroup

	for range count {
		wg.Go(func() {
			ids <- uniqueID()
		})
	}
	wg.Wait()
	close(ids)

	seen := make(map[int64]bool, count)
	for id := range ids {
		assert.False(t, seen[id], "duplicate toast ID %d", id)
		seen[id] = true
	}
	require.Len(t, seen, count)
}

func TestToast_RenderedDismissExpressionsUseGeneratedSafeID(t *testing.T) {
	idCounter.Store(0)
	rendered := renderToast(t, Config{
		Tone:    ToneSuccess,
		Title:   "Saved",
		Message: "Done",
	})
	browserHTML := html.UnescapeString(rendered)

	assert.Contains(t, browserHTML, `id="server-toast-1"`)
	assert.Contains(t, browserHTML, `data-toast-id="server-toast-1"`)
	assert.Contains(t, browserHTML, `x-on:toast-dismiss.window="if ($event.detail.id === 'server-toast-1') { $el.remove() }"`)
	assert.Contains(t, browserHTML, `x-on:click="isVisible = false; $dispatch('toast-dismiss', { id: 'server-toast-1' })"`)
}

func TestSingleToastAlpineData_UsesNumericDuration(t *testing.T) {
	data := singleToastAlpineData(1250, false)

	assert.Contains(t, data, "}, 1250);")
	assert.NotContains(t, data, "}, '1250');")
}

func TestSingleToastAlpineData_PersistentHasNoAutoDismiss(t *testing.T) {
	data := singleToastAlpineData(1250, true)

	assert.Contains(t, data, "isVisible: true")
	assert.NotContains(t, data, "setTimeout")
	assert.NotContains(t, data, "toast-dismiss")
}
