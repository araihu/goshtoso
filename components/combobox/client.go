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
// The listener itself (emitted by the private client script fragment) is safe to render repeatedly —
// a module-init guard (window.__goshtosoComboboxInit) prevents double-binding.
// The client script is emitted automatically by Combobox when
// Source.LazyEndpoint is empty; consumers rarely need to call it directly.
const ClientEvent = "combobox:change"
