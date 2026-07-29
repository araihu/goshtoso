package combobox

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
// Client-mode behavior lives in the reusable Goshtoso component-runtime bundle.
// Consumers load that bundle through head.Dependencies or head.DependenciesMinimal.
const ClientEvent = "combobox:change"
