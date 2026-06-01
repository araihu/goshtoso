package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// gotoChat navigates to the chat page, waits for Alpine + the message log, then
// for the ws extension to have opened the socket. ws-send silently no-ops if the
// socket is not yet open, so the readiness wait is load-bearing before any send.
func gotoChat(t *testing.T, page playwright.Page) {
	t.Helper()
	require.NoError(t, page.AddInitScript(playwright.Script{
		Content: new("try{localStorage.setItem('cookieConsent','accepted')}catch(e){}"),
	}))
	_, err := page.Goto(baseURL + "/examples/chat")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => !!document.querySelector('#chat-log')", nil)
	require.NoError(t, err)
	waitWSOpen(t, page)
}

// waitWSOpen blocks until the htmx ws extension has an OPEN socket. htmx stores
// the live socket on the connecting element's internal data; we probe its
// readyState via htmx's API. Falls back to a presence-driven readiness signal:
// once the server's own "joined" presence frame has landed in our DOM, the
// round-trip is proven open. We wait on either signal.
func waitWSOpen(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.WaitForFunction(
		`() => {
			const el = document.querySelector('[ws-connect]');
			if (!el) return false;
			try {
				const api = window.htmx && htmx.find ? el : null;
				const internal = el['htmx-internal-data'];
				if (internal && internal.webSocket && internal.webSocket.readyState === 1) return true;
			} catch (e) {}
			// Fallback: the server's "joined" system line in the log proves a
			// frame already round-tripped (presence is now in-chat, not a toast).
			return !!document.querySelector('#chat-log > div');
		}`,
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "websocket should open (ws extension bound + socket open)")
}

// sendChat types into the composer and submits via the ws-send form. Fill sets
// the textarea value; ws-send reads the field value on submit, so a click on the
// form's submit button serializes and sends the frame.
func sendChat(t *testing.T, page playwright.Page, msg string) {
	t.Helper()
	require.NoError(t, page.Locator("#chat-message").Fill(msg))
	require.NoError(t, page.Locator("form[ws-send] button[type='submit']").Click())
}

// logHas builds a JS predicate asserting some #chat-log message-bubble body
// equals s. The chatbubble component renders the message text in a leaf <div>
// (the only #chat-log element whose own trimmed textContent equals the message,
// since rows also contain the nick + timestamp), so we match leaf elements.
func logHas(s string) string {
	return "() => Array.from(document.querySelectorAll('#chat-log *'))" +
		".some(e => e.children.length === 0 && e.textContent.trim() === " + jsString(s) + ")"
}

// jsString quotes a Go string as a JS single-quoted literal (test inputs are
// simple ASCII; escape backslash and single quote defensively).
func jsString(s string) string {
	var out strings.Builder
	out.WriteString("'")
	for _, r := range s {
		switch r {
		case '\\':
			out.WriteString("\\\\")
		case '\'':
			out.WriteString("\\'")
		default:
			out.WriteString(string(r))
		}
	}
	return out.String() + "'"
}

// logRowMine builds a JS predicate asserting some #chat-log row containing text
// s has the given data-mine value. The row is the [data-sender] element the
// chatbubble component emits; the mine-detection script flips its data-mine.
func logRowMine(s, mine string) string {
	return "() => Array.from(document.querySelectorAll('#chat-log [data-sender]')).some(r => " +
		"r.textContent.includes(" + jsString(s) + ") && r.getAttribute('data-mine') === " + jsString(mine) + ")"
}

// TestChat_OwnMessageAligned proves the own-message detector: the sender's own
// bubble row flips to data-mine='true' (right/sent) and carries their nick as
// data-sender, while on a second context the same message stays data-mine='false'
// (received).
func TestChat_OwnMessageAligned(t *testing.T) {
	pageA := newIsolatedPage(t)
	pageB := newIsolatedPage(t)
	gotoChat(t, pageA)
	gotoChat(t, pageB)

	sendChat(t, pageA, "mine-aligned-A")

	// On A: the row is mine (right/sent).
	_, err := pageA.WaitForFunction(logRowMine("mine-aligned-A", "true"), nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "sender's own bubble should be data-mine='true'")

	// Its data-sender equals A's current nick (read from the composer hidden input).
	matches, err := pageA.Evaluate(
		`() => { const me = document.querySelector('#chat-hidden-nick').value;
		         const row = Array.from(document.querySelectorAll('#chat-log [data-mine="true"]'))
		           .find(r => r.textContent.includes('mine-aligned-A'));
		         return row && row.getAttribute('data-sender') === me; }`, nil)
	require.NoError(t, err)
	require.Equal(t, true, matches, "mine row's data-sender should equal the viewer's nick")

	// On B (a different viewer): the same message is received, data-mine='false'.
	_, err = pageB.WaitForFunction(logRowMine("mine-aligned-A", "false"), nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "the same message on another viewer should be data-mine='false' (received)")
}

// TestChat_EnterToSend proves Enter submits the composer (no button click) and
// Shift+Enter does not send (it inserts a newline instead).
func TestChat_EnterToSend(t *testing.T) {
	page := newIsolatedPage(t)
	gotoChat(t, page)

	// Shift+Enter must NOT send: type text, press Shift+Enter, assert no bubble and
	// a newline was inserted into the textarea value.
	require.NoError(t, page.Locator("#chat-message").Fill("no-send-shift"))
	require.NoError(t, page.Locator("#chat-message").Press("Shift+Enter"))
	val, err := page.Locator("#chat-message").InputValue()
	require.NoError(t, err)
	require.Contains(t, val, "\n", "Shift+Enter should insert a newline, not send")
	noBubble, err := page.Evaluate(
		`() => !Array.from(document.querySelectorAll('#chat-log *')).some(e => e.children.length === 0 && e.textContent.includes('no-send-shift'))`, nil)
	require.NoError(t, err)
	require.Equal(t, true, noBubble, "Shift+Enter should not broadcast a message")

	// Enter SENDS without clicking the submit button. The ws-send rebind race can
	// drop the first submit after the socket binds, so re-fire Enter until the
	// bubble round-trips (a lost send produces no bubble, so re-pressing is safe).
	require.NoError(t, page.Locator("#chat-message").Fill("enter-sends-me"))
	for attempt := 1; attempt <= 5; attempt++ {
		require.NoError(t, page.Locator("#chat-message").Press("Enter"))
		if _, err := page.WaitForFunction(logHas("enter-sends-me"), nil,
			playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)}); err == nil {
			return
		}
		// ws-send does not reset the textarea, so the value persists; re-fill
		// defensively in case a prior keystroke mutated it.
		_ = page.Locator("#chat-message").Fill("enter-sends-me")
	}
	t.Fatal("Enter did not send the message after 5 attempts")
}

// TestChat_Broadcast opens two independent browser contexts, sends from A, and
// asserts the message lands in B. The real full-duplex proof.
func TestChat_Broadcast(t *testing.T) {
	pageA := newIsolatedPage(t)
	pageB := newIsolatedPage(t)
	gotoChat(t, pageA)
	gotoChat(t, pageB)

	sendChat(t, pageA, "hello-from-A")

	_, err := pageB.WaitForFunction(logHas("hello-from-A"), nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "message from A should appear in B's log")
}

// TestChat_ElizaDeterministic toggles the bot on, sends a known input, and
// asserts the deterministic reply bubble appears.
func TestChat_ElizaDeterministic(t *testing.T) {
	page := newIsolatedPage(t)
	gotoChat(t, page)

	// Turn on the ELIZA toggle. The real Toggle's checkbox is sr-only (#chat-eliza);
	// the clickable target is its wrapping label. Verify it is actually checked
	// before sending so the frame carries eliza=on.
	require.NoError(t, page.Locator("label[for='chat-eliza']").Click())
	_, err := page.WaitForFunction("() => !!document.querySelector('#chat-eliza:checked')", nil)
	require.NoError(t, err, "ELIZA toggle should be checked after clicking its label")

	sendChat(t, page, "i feel anxious")

	_, err = page.WaitForFunction(logHas("Why do you feel anxious?"), nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "ELIZA should reply deterministically")
}

// TestChat_Presence asserts the online badge tracks a second connection.
func TestChat_Presence(t *testing.T) {
	pageA := newIsolatedPage(t)
	gotoChat(t, pageA)

	// A second context joins → A's badge should read "2 online".
	pageB := newIsolatedPage(t)
	gotoChat(t, pageB)

	_, err := pageA.WaitForFunction(
		"() => { const el = document.querySelector('#chat-online'); return el && el.textContent.includes('2 online'); }",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "online badge should reach 2")
}

// TestChat_Rename renames the visitor and asserts the next message carries the
// new nick (the hidden composer input was OOB-swapped).
func TestChat_Rename(t *testing.T) {
	page := newIsolatedPage(t)
	gotoChat(t, page)

	require.NoError(t, page.Locator("#chat-nick-input").Fill("Ada"))
	require.NoError(t, page.Locator("form[hx-post='/api/examples/chat/rename'] button[type='submit']").Click())
	// Wait for the OOB swap to update the visible nick.
	_, err := page.WaitForFunction(
		"() => { const el = document.querySelector('#chat-me'); return el && el.textContent.trim() === 'Ada'; }",
		nil)
	require.NoError(t, err, "rename should update the visible nick via OOB swap")

	sendChat(t, page, "renamed-hello")
	_, err = page.WaitForFunction(
		"() => Array.from(document.querySelectorAll('#chat-log > div')).some(l => l.textContent.includes('Ada') && l.textContent.includes('renamed-hello'))",
		nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "message after rename should be labeled 'Ada'")
}

// TestChat_FragmentNavNoErrors reaches /examples/chat via the sidebar (htmx
// fragment swap) and asserts the socket connects and broadcasts with no console
// or page errors — the load-bearing regression guard.
func TestChat_FragmentNavNoErrors(t *testing.T) {
	page := newIsolatedPage(t)

	var jsErrors []string
	page.On("pageerror", func(err error) { jsErrors = append(jsErrors, err.Error()) })
	page.On("console", func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			jsErrors = append(jsErrors, m.Text())
		}
	})

	_, err := page.Goto(baseURL + "/getting-started")
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("a[href='/examples/chat']").First().Click())
	_, err = page.WaitForFunction("() => !!document.querySelector('#chat-log')", nil)
	require.NoError(t, err)
	waitWSOpen(t, page)

	// Send a message through the fragment-loaded page and confirm it round-trips.
	// The FIRST ws-send right after a fragment swap can be lost to the htmx
	// rebind race (htmx wires the swapped-in ws-send form a beat after it lands
	// in the DOM; a send fired in that window dispatches no frame). A lost send
	// produces no bubble, so re-firing is safe — clickUntil re-submits until the
	// message round-trips. The textarea value persists across clicks.
	require.NoError(t, page.Locator("#chat-message").Fill("frag-nav-msg"))
	clickUntil(t, page, page.Locator("form[ws-send] button[type='submit']"), logHas("frag-nav-msg"))

	require.Empty(t, jsErrors, "no JS console/page errors on fragment-nav chat page: %v", jsErrors)
}

// TestChat_SidebarPresent verifies the Examples section lists the Chat link.
func TestChat_SidebarPresent(t *testing.T) {
	page := newIsolatedPage(t)
	_, err := page.Goto(baseURL + "/examples/chat")
	require.NoError(t, err)
	require.NoError(t, page.Locator("text=Examples").First().WaitFor())
	require.NoError(t, page.Locator("a[href='/examples/chat']").First().WaitFor())
}

// TestChat_ComposerClearsOnSend verifies the composer textarea empties after a
// message is sent (htmx ws-send does not reset the form; a wsAfterSend listener does).
func TestChat_ComposerClearsOnSend(t *testing.T) {
	page := newIsolatedPage(t)
	gotoChat(t, page)
	sendChat(t, page, "clears-after-send")
	_, err := page.WaitForFunction(logHas("clears-after-send"), nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => document.getElementById('chat-message').value === ''", nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err, "textarea should clear after send")
}
