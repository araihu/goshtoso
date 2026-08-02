// Package chatpage owns the Chat runnable example page.
package chatpage

import (
	"github.com/a-h/templ"
	chatdomain "github.com/araihu/goshtoso/site/internal/examples/chat"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
)

// Definition is the Chat example's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "examples/chat",
	Title:   "Chat",
	Active:  "chat",
	Type:    "SoftwareSourceCode",
	Content: func() templ.Component { return ChatApp(chatdomain.NewGuest(0)) },
}
