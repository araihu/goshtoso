package e2e

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/form"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestFormNavigationSinksRejectExecutableSchemes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("submit action", func(t *testing.T) {
		page := newPage(t, sharedBrowser)
		dialogSeen := listenForDialogs(t, page)

		html := renderComponentDocument(t, form.Form(form.Config{
			Action: "javascript:alert('form-action-xss')",
			Footer: &form.FooterConfig{
				SubmitLabel: "Submit",
			},
		}))
		require.NoError(t, page.SetContent(html))

		require.NoError(t, page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Submit",
		}).Click())
		requireNoDialog(t, dialogSeen, "submit executed javascript: form action")
	})

	t.Run("cancel href", func(t *testing.T) {
		page := newPage(t, sharedBrowser)
		dialogSeen := listenForDialogs(t, page)

		html := renderComponentDocument(t, form.Form(form.Config{
			Footer: &form.FooterConfig{
				CancelLabel: "Cancel",
				CancelHref:  "javascript:alert('form-cancel-xss')",
				SubmitLabel: "Submit",
			},
		}))
		require.NoError(t, page.SetContent(html))

		require.NoError(t, page.GetByRole("link", playwright.PageGetByRoleOptions{
			Name: "Cancel",
		}).Click())
		requireNoDialog(t, dialogSeen, "cancel link executed javascript: href")
	})
}

func renderComponentDocument(t *testing.T, c templ.Component) string {
	t.Helper()

	return "<!doctype html><html><body>" + renderComponentFragment(t, c) + "</body></html>"
}

func renderComponentFragment(t *testing.T, c templ.Component) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, c.Render(context.Background(), &buf))

	return buf.String()
}

func listenForDialogs(t *testing.T, page playwright.Page) <-chan string {
	t.Helper()

	dialogSeen := make(chan string, 1)
	page.On("dialog", func(dialog playwright.Dialog) {
		select {
		case dialogSeen <- dialog.Message():
		default:
		}
		require.NoError(t, dialog.Accept())
	})
	return dialogSeen
}

func requireNoDialog(t *testing.T, dialogSeen <-chan string, context string) {
	t.Helper()

	select {
	case msg := <-dialogSeen:
		t.Fatalf("%s: %s", context, msg)
	case <-time.After(300 * time.Millisecond):
	}
}
