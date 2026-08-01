// Package chatbubblepage owns the Chat Bubble component documentation page.
package chatbubblepage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Chat Bubble page's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "components/chatbubble",
	Title:   "Chat Bubble",
	Active:  "chatbubble",
	Type:    "TechArticle",
	Content: chatBubbleDemoContent,
}
