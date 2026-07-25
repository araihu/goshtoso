package components

import (
	"context"
	"reflect"
	"strings"
	"testing"

	rootcomponents "github.com/araihu/goshtoso/components"
	"github.com/araihu/goshtoso/components/accordion"
	"github.com/araihu/goshtoso/components/avatar"
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/banner"
	"github.com/araihu/goshtoso/components/card"
	"github.com/araihu/goshtoso/components/carousel"
	"github.com/araihu/goshtoso/components/chatbubble"
	"github.com/araihu/goshtoso/components/codeblock"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/stretchr/testify/require"
)

func TestDisplayAPIMetadataRegistered(t *testing.T) {
	expected := map[string][]rootcomponents.Kind{
		"components/accordion": {
			rootcomponents.KindAccordion,
		},
		"components/avatar": {
			rootcomponents.KindAvatar,
			rootcomponents.KindAvatarStack,
		},
		"components/badge": {
			rootcomponents.KindBadge,
			rootcomponents.KindNotificationBadge,
			rootcomponents.KindNotificationDot,
			rootcomponents.KindAnimatingDot,
		},
		"components/banner": {
			rootcomponents.KindBanner,
			rootcomponents.KindCookieBanner,
		},
		"components/card": {
			rootcomponents.KindCard,
		},
		"components/carousel": {
			rootcomponents.KindCarousel,
			rootcomponents.KindCardCarousel,
		},
		"components/chatbubble": {
			rootcomponents.KindChatBubble,
			rootcomponents.KindTypingIndicator,
		},
		"components/codeblock": {
			rootcomponents.KindCodeBlock,
		},
		"components/dependencies": {
			rootcomponents.KindDependencies,
			rootcomponents.KindDependenciesMinimal,
		},
		"components/kbd": {
			rootcomponents.KindKbd,
		},
		"components/table": {
			rootcomponents.KindTable,
			rootcomponents.KindTableHeadContent,
			rootcomponents.KindTableRows,
			rootcomponents.KindTableRow,
			rootcomponents.KindTablePaginationNav,
			rootcomponents.KindTableImageCell,
		},
	}

	require.Len(t, expected, 11)
	displayCatalogEntries := make([]string, 0, len(expected))
	for _, page := range catalog.ComponentPages() {
		if page.Section == "Display" {
			displayCatalogEntries = append(displayCatalogEntries, page.Key)
		}
	}
	require.Len(t, displayCatalogEntries, 11)
	for _, key := range displayCatalogEntries {
		require.Containsf(t, expected, key, "unexpected Display catalog entry %q", key)
	}

	for key, wantKinds := range expected {
		entry, ok := Demos[key]
		require.Truef(t, ok, "missing Display registry entry %q", key)
		require.NotEmptyf(t, entry.API, "%s must register structured API metadata", key)

		seenIDs := make(map[string]struct{}, len(entry.API))
		gotKinds := make([]rootcomponents.Kind, 0, len(entry.API))
		for _, section := range entry.API {
			require.NotEmptyf(t, section.ID, "%s contains an empty API section ID", key)
			_, duplicate := seenIDs[section.ID]
			require.Falsef(t, duplicate, "%s contains duplicate API section ID %q", key, section.ID)
			seenIDs[section.ID] = struct{}{}
			if section.Kind != "" {
				gotKinds = append(gotKinds, section.Kind)
			}
		}
		require.ElementsMatchf(t, wantKinds, gotKinds, "%s API section Kinds", key)
	}
}

func TestDisplayAPIRegistryUsesPageSectionSlices(t *testing.T) {
	expected := map[string][]demo.APISection{
		"components/accordion":    accordionAPISections,
		"components/avatar":       avatarAPISections,
		"components/badge":        badgeAPISections,
		"components/banner":       bannerAPISections,
		"components/card":         cardAPISections,
		"components/carousel":     carouselAPISections,
		"components/chatbubble":   chatBubbleAPISections,
		"components/codeblock":    codeBlockAPISections,
		"components/dependencies": dependenciesAPISections,
		"components/kbd":          kbdAPISections,
		"components/table":        tableAPISections,
	}

	require.Len(t, expected, 11)
	for key, pageSections := range expected {
		entry := Demos[key]
		require.NotEmpty(t, entry.API, key)
		require.NotEmpty(t, pageSections, key)
		require.Samef(
			t,
			&pageSections[0],
			&entry.API[0],
			"%s registry API must use the page's named metadata slice",
			key,
		)

		var rendered strings.Builder
		require.NoError(t, entry.Content().Render(context.Background(), &rendered))
		html := rendered.String()
		require.Equalf(t, 1, strings.Count(html, "data-api-reference"), "%s must render one structured API reference", key)
		for _, section := range pageSections {
			require.Equalf(
				t,
				1,
				strings.Count(html, `data-api-section="`+section.ID+`"`),
				"%s must render its %q API section exactly once",
				key,
				section.ID,
			)
		}
	}
}

func TestDisplayStructAPIsDocumentEveryExportedFieldExactlyOnce(t *testing.T) {
	expectedTypes := []reflect.Type{
		reflect.TypeFor[accordion.AccordionConfig](),
		reflect.TypeFor[accordion.AccordionItem](),
		reflect.TypeFor[avatar.Config](),
		reflect.TypeFor[avatar.StackConfig](),
		reflect.TypeFor[badge.Config](),
		reflect.TypeFor[banner.Config](),
		reflect.TypeFor[banner.CTAConfig](),
		reflect.TypeFor[banner.CookieBannerConfig](),
		reflect.TypeFor[card.Config](),
		reflect.TypeFor[carousel.Config](),
		reflect.TypeFor[carousel.CardConfig](),
		reflect.TypeFor[carousel.Slide](),
		reflect.TypeFor[carousel.AutoplayConfig](),
		reflect.TypeFor[carousel.HTMXConfig](),
		reflect.TypeFor[chatbubble.Config](),
		reflect.TypeFor[codeblock.Config](),
		reflect.TypeFor[table.Config](),
		reflect.TypeFor[table.Column](),
		reflect.TypeFor[table.Cell](),
		reflect.TypeFor[table.Row](),
		reflect.TypeFor[table.RowHTMXConfig](),
		reflect.TypeFor[table.PaginationConfig](),
		reflect.TypeFor[table.InfiniteScrollConfig](),
		reflect.TypeFor[table.FilterConfig](),
		reflect.TypeFor[table.Filter](),
		reflect.TypeFor[table.FilterOption](),
		reflect.TypeFor[table.FilterOptionsHTMXConfig](),
		reflect.TypeFor[table.FilterHTMXConfig](),
		reflect.TypeFor[table.HTMXConfig](),
	}

	sectionsByType := make(map[reflect.Type][]demo.APISection, len(expectedTypes))
	for _, sections := range displayAPISectionSlices() {
		for _, section := range sections {
			require.NotEmpty(t, section.ID)
			require.NotEmpty(t, section.Title)
			require.NotEmptyf(t, section.Description, "%s section description", section.Title)
			for _, prop := range section.Props {
				require.NotEmptyf(t, prop.Default, "%s.%s default", section.Title, prop.Name)
				require.NotEmptyf(t, prop.Description, "%s.%s description", section.Title, prop.Name)
			}
			if section.StructType != nil {
				sectionsByType[section.StructType] = append(sectionsByType[section.StructType], section)
			}
		}
	}
	require.Len(t, sectionsByType, len(expectedTypes))

	for _, typ := range expectedTypes {
		sections := sectionsByType[typ]
		require.Lenf(t, sections, 1, "%s must have exactly one StructAPI section", typ)
		section := sections[0]

		wantFields := make([]string, 0, typ.NumField())
		for index := range typ.NumField() {
			field := typ.Field(index)
			if field.IsExported() {
				wantFields = append(wantFields, field.Name)
			}
		}
		gotFields := make([]string, 0, len(section.Props))
		for _, prop := range section.Props {
			gotFields = append(gotFields, prop.Name)
		}
		require.ElementsMatchf(t, wantFields, gotFields, "%s exported fields", typ)
		require.Lenf(t, gotFields, len(wantFields), "%s must document every field once", typ)
	}
}

func TestDisplayMetadataCapturesRenderedDefaultsAndPublicSignatures(t *testing.T) {
	require.Equal(t, `"table"`, apiProp(t, tableAPISections, "Config", "ID").Default)
	require.Equal(t, `"load"`, apiProp(t, tableAPISections, "Config", "LazyTrigger").Default)
	require.Equal(t, []string{"AppearanceDefault", "AppearanceStriped"}, apiProp(t, tableAPISections, "Config", "Appearance").Allowed)
	require.Contains(t, apiProp(t, tableAPISections, "Cell", "BadgeColor").Description, "badge.Badge")
	require.Equal(t, `"innerHTML"`, apiProp(t, tableAPISections, "RowHTMXConfig", "Swap").Default)
	require.Equal(t, "false when collapsible; otherwise visible", apiProp(t, tableAPISections, "FilterConfig", "InitiallyExpanded").Default)
	require.Equal(t, `"#<table-id>-tbody"`, apiProp(t, tableAPISections, "FilterHTMXConfig", "Target").Default)
	require.Equal(t, `"innerHTML"`, apiProp(t, tableAPISections, "FilterHTMXConfig", "Swap").Default)

	require.Equal(
		t,
		"func (cfg Config) NextSortDir(key string) SortDir",
		apiProp(t, tableAPISections, "Config helpers", "NextSortDir").Signature,
	)
	require.Equal(
		t,
		"func ImageCell(imageURL string, label string, detail string) ImageCellInstance",
		apiProp(t, tableAPISections, "ImageCell", "ImageCell").Signature,
	)

	require.Equal(t, "4000ms when enabled", apiProp(t, carouselAPISections, "AutoplayConfig", "Interval").Default)
	require.Equal(t, `"load"`, apiProp(t, carouselAPISections, "HTMXConfig", "Trigger").Default)
	require.Equal(t, `"innerHTML"`, apiProp(t, carouselAPISections, "HTMXConfig", "Swap").Default)

	require.Equal(t, `"Cookie Consent"`, apiProp(t, bannerAPISections, "CookieBannerConfig", "Title").Default)
	require.Equal(t, `"Accept"`, apiProp(t, bannerAPISections, "CookieBannerConfig", "AcceptLabel").Default)
	require.Equal(t, `"Decline"`, apiProp(t, bannerAPISections, "CookieBannerConfig", "RejectLabel").Default)

	require.Equal(
		t,
		"func NotificationBadge(count int) NotificationBadgeInstance",
		apiProp(t, badgeAPISections, "NotificationBadge", "NotificationBadge").Signature,
	)
	require.Equal(t, "count <= 0 renders nothing; counts > 99 render 99+", apiProp(t, badgeAPISections, "NotificationBadge", "NotificationBadge").Default)

	require.Equal(t, "SizeMD", apiProp(t, kbdAPISections, "Kbd options", "WithSize").Default)
	require.Equal(
		t,
		"func WithAttrs(attrs templ.Attributes) Option",
		apiProp(t, kbdAPISections, "Kbd options", "WithAttrs").Signature,
	)
}

func TestAccordionMetadataExplainsItemIDNamespace(t *testing.T) {
	configID := apiProp(t, accordionAPISections, "AccordionConfig", "ID")
	require.Contains(t, configID.Description, "root element")
	require.Contains(t, configID.Description, "does not namespace")

	itemID := apiProp(t, accordionAPISections, "AccordionItem", "ID")
	require.Contains(t, itemID.Description, "namespace source")
	require.Contains(t, itemID.Description, "multiple accordions")
}

func TestCarouselMetadataMarksStaticOnlyFields(t *testing.T) {
	for _, name := range []string{"Autoplay", "Touch", "AspectRatio", "Height"} {
		description := apiProp(t, carouselAPISections, "Config", name).Description
		require.Containsf(t, description, "static mode", "Config.%s", name)
		require.Containsf(t, description, "ignored when HTMX is non-nil", "Config.%s", name)
	}
}

func TestTableMetadataDocumentsConditionalOptionalFields(t *testing.T) {
	totalPages := apiProp(t, tableAPISections, "PaginationConfig", "TotalPages")
	require.False(t, totalPages.Required)
	require.Equal(t, "0 (no numbered navigation)", totalPages.Default)
	require.Contains(t, totalPages.Description, "traditional mode")
	require.Contains(t, totalPages.Description, "ignored in infinite-scroll mode")

	optionValue := apiProp(t, tableAPISections, "FilterOption", "Value")
	require.False(t, optionValue.Required)
	require.Equal(t, `""`, optionValue.Default)
	require.Contains(t, optionValue.Description, "empty string is valid")
}

func displayAPISectionSlices() [][]demo.APISection {
	return [][]demo.APISection{
		accordionAPISections,
		avatarAPISections,
		badgeAPISections,
		bannerAPISections,
		cardAPISections,
		carouselAPISections,
		chatBubbleAPISections,
		codeBlockAPISections,
		dependenciesAPISections,
		kbdAPISections,
		tableAPISections,
	}
}

func apiProp(
	t *testing.T,
	sections []demo.APISection,
	sectionTitle string,
	propName string,
) demo.APIPropDoc {
	t.Helper()
	for _, section := range sections {
		if section.Title != sectionTitle {
			continue
		}
		for _, prop := range section.Props {
			if prop.Name == propName {
				return prop
			}
		}
		require.FailNowf(t, "missing API prop", "%s.%s", sectionTitle, propName)
	}
	require.FailNowf(t, "missing API section", "%s", sectionTitle)
	return demo.APIPropDoc{}
}
