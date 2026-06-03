package combobox

import _ "embed"

//go:embed client.js
var clientJS string

// ClientEvent is the CustomEvent name dispatched by the client-side listener
// when the combobox selection changes. The event bubbles from the combobox root
// element. Event.detail has the shape:
//
//	{ id: string, values: string[] }
//
// where id is the combobox cfg.ID and values is the current selected set.
// Parent pages listen for this event to trigger form submission or reactive UI
// updates.
//
// The listener itself (emitted by ClientScript) is safe to render repeatedly —
// a module-init guard (window.__goshtosoComboboxV2Init) prevents double-binding.
// ClientScript is emitted automatically by Combobox when cfg.IsClientMode() is
// true; consumers rarely need to call it directly.
const ClientEvent = "combobox:change"
