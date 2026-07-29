package components_test

import (
	"testing"

	"github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/accordion"
	"github.com/araihu/goshtoso/components/avatar"
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/banner"
	"github.com/araihu/goshtoso/components/card"
	"github.com/araihu/goshtoso/components/carousel"
	"github.com/araihu/goshtoso/components/chatbubble"
	"github.com/araihu/goshtoso/components/codeblock"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/icon"
	"github.com/araihu/goshtoso/components/kbd"
	"github.com/araihu/goshtoso/components/table"
	"github.com/stretchr/testify/require"
)

func displayRenderables() map[components.Kind]components.Component {
	return map[components.Kind]components.Component{
		components.KindAccordion:           accordion.Accordion(accordion.AccordionConfig{}),
		components.KindAvatar:              avatar.Avatar(avatar.Config{}),
		components.KindAvatarStack:         avatar.AvatarStack(avatar.StackConfig{}),
		components.KindBadge:               badge.Badge(badge.Config{}),
		components.KindNotificationBadge:   badge.NotificationBadge(0),
		components.KindNotificationDot:     badge.NotificationDot(),
		components.KindAnimatingDot:        badge.AnimatingDot(badge.ToneDefault),
		components.KindBanner:              banner.Banner(banner.Config{}),
		components.KindCookieBanner:        banner.CookieBanner(banner.CookieBannerConfig{}),
		components.KindCard:                card.Card(card.Config{}),
		components.KindCarousel:            carousel.Carousel(carousel.Config{}),
		components.KindCardCarousel:        carousel.CardCarousel(carousel.CardConfig{}),
		components.KindChatBubble:          chatbubble.ChatBubble(chatbubble.Config{}),
		components.KindTypingIndicator:     chatbubble.TypingIndicator(chatbubble.Config{}),
		components.KindCodeBlock:           codeblock.CodeBlock(codeblock.Config{}),
		components.KindDependencies:        head.Dependencies(),
		components.KindDependenciesMinimal: head.DependenciesMinimal(),
		components.KindIcon:                icon.Icon(icon.Config{SpriteURL: "/sprites/ui.svg", Symbol: "check"}),
		components.KindKbd:                 kbd.Kbd(""),
		components.KindTable:               table.Table(table.Config{}),
		components.KindTableHeadContent:    table.TableHeadContent(table.Config{}),
		components.KindTableRows:           table.TableRows(table.Config{}),
		components.KindTableRow:            table.TableRow(table.Config{}, table.Row{}),
		components.KindTablePaginationNav:  table.TablePaginationNav(table.Config{}),
		components.KindTableImageCell:      table.ImageCell("", "", ""),
	}
}

func TestDisplayRenderablesExposeKinds(t *testing.T) {
	values := displayRenderables()
	require.Len(t, values, 25)
	for want, value := range values {
		require.Equal(t, want, value.Kind())
	}
}
