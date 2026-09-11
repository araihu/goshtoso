# Built-in expression inventory

Reconciled against origin/main commit `8bfeca12` and the component audit of
`5256b101`. These locations identify authored implementation source in this feature.
Generated templ output and JS bundles are excluded. Existing explicit component
fields retain precedence over the expression set.

| Expression | English default / message | Source locations |
| --- | --- | --- |
| `ActionGroup.AriaLabel` | Actions | [components/actiongroup/expressions.go:12](../components/actiongroup/expressions.go#L12) |
| `ActionGroup.OverflowAriaLabel` | More actions | [components/actiongroup/expressions.go:13](../components/actiongroup/expressions.go#L13) |
| `Alert.DismissAriaLabel` | dismiss alert | [components/alert/alert.templ:54](../components/alert/alert.templ#L54) |
| `Alert.DismissLabel` | Dismiss | [components/alert/alert.templ:82](../components/alert/alert.templ#L82) |
| `AppShell.SkipLinkLabel` | Skip to main content | [components/appshell/appshell.templ:17](../components/appshell/appshell.templ#L17) |
| `Avatar.GroupAriaLabel` | Avatar group | [components/avatar/avatar.templ:50](../components/avatar/avatar.templ#L50) |
| `Badge.NotificationAriaLabel` | notification | [components/badge/badge.templ:108](../components/badge/badge.templ#L108) |
| `Banner.DismissAriaLabel` | dismiss banner | [components/banner/banner.templ:79](../components/banner/banner.templ#L79) |
| `Banner.ConsentAriaLabel` | Cookie consent | [components/banner/banner.templ:99](../components/banner/banner.templ#L99) |
| `Banner.Title` | Cookie Consent | [components/banner/banner.templ:109](../components/banner/banner.templ#L109) |
| `Banner.AcceptLabel` | Accept | [components/banner/banner.templ:138](../components/banner/banner.templ#L138) |
| `Banner.RejectLabel` | Decline | [components/banner/banner.templ:128](../components/banner/banner.templ#L128) |
| `Breadcrumbs.AriaLabel` | breadcrumb | [components/breadcrumbs/breadcrumbs.templ:18](../components/breadcrumbs/breadcrumbs.templ#L18) |
| `Carousel.LoadingText` | Loading carousel... | [components/carousel/carousel.templ:176](../components/carousel/carousel.templ#L176) |
| `Carousel.PreviousAriaLabel` | previous slide | [components/carousel/carousel.templ:186](../components/carousel/carousel.templ#L186) |
| `Carousel.NextAriaLabel` | next slide | [components/carousel/carousel.templ:200](../components/carousel/carousel.templ#L200) |
| `Carousel.SlidesAriaLabel` | slides | [components/carousel/carousel.templ:219](../components/carousel/carousel.templ#L219) |
| `Carousel.PauseAriaLabel` | pause carousel | [components/carousel/carousel.templ:238](../components/carousel/carousel.templ#L238), [components/carousel/carousel.templ:239](../components/carousel/carousel.templ#L239) |
| `Carousel.PlayAriaLabel` | play carousel | [components/carousel/carousel.templ:240](../components/carousel/carousel.templ#L240) |
| `Carousel.SlideLabel` | Complete application-formatted message callback | [components/carousel/carousel.templ:220](../components/carousel/carousel.templ#L220) |
| `ChatBubble.TypingAriaLabel` | typing | [components/chatbubble/chatbubble.templ:70](../components/chatbubble/chatbubble.templ#L70) |
| `ChatBubble.SendingLabel` | Sending | [components/chatbubble/expressions.go:12](../components/chatbubble/expressions.go#L12) |
| `ChatBubble.DeliveredLabel` | Delivered | [components/chatbubble/expressions.go:14](../components/chatbubble/expressions.go#L14) |
| `ChatBubble.SeenLabel` | Seen | [components/chatbubble/expressions.go:16](../components/chatbubble/expressions.go#L16) |
| `ChatBubble.BotLabel` | BOT | [components/chatbubble/chatbubble.templ:108](../components/chatbubble/chatbubble.templ#L108) |
| `CodeBlock.CopyLabel` | Copy | [components/codeblock/codeblock.templ:42](../components/codeblock/codeblock.templ#L42), [components/codeblock/codeblock.templ:54](../components/codeblock/codeblock.templ#L54) |
| `CodeBlock.CopiedLabel` | Copied! | [components/codeblock/codeblock.templ:43](../components/codeblock/codeblock.templ#L43) |
| `CodeBlock.ErrorText` | Unable to copy | [components/codeblock/codeblock.templ:44](../components/codeblock/codeblock.templ#L44) |
| `CodeBlock.CopyAriaLabel` | Complete application-formatted message callback | [components/codeblock/codeblock.templ:41](../components/codeblock/codeblock.templ#L41) |
| `Combobox.Placeholder` | Select… | [components/combobox/expressions.go:11](../components/combobox/expressions.go#L11) |
| `Combobox.ClearLabel` | Clear all | [components/combobox/combobox.templ:61](../components/combobox/combobox.templ#L61), [components/combobox/combobox.templ:73](../components/combobox/combobox.templ#L73) |
| `Combobox.SearchPlaceholder` | Search… | [components/combobox/combobox.templ:85](../components/combobox/combobox.templ#L85) |
| `Combobox.EmptyText` | No matches found | [components/combobox/combobox.templ:112](../components/combobox/combobox.templ#L112) |
| `Combobox.ErrorText` | Failed to load. | [components/combobox/error.templ:11](../components/combobox/error.templ#L11) |
| `Combobox.RetryLabel` | Retry | [components/combobox/error.templ:12](../components/combobox/error.templ#L12) |
| `Combobox.SelectedLabel` | Complete application-formatted message callback | [components/combobox/expressions.go:17](../components/combobox/expressions.go#L17), [components/combobox/expressions.go:23](../components/combobox/expressions.go#L23) |
| `Drawer.CloseAriaLabel` | Close | [components/drawer/drawer.templ:85](../components/drawer/drawer.templ#L85) |
| `Dropdown.ContextMenuAriaLabel` | context menu | [components/dropdown/dropdown.templ:156](../components/dropdown/dropdown.templ#L156) |
| `Dropdown.OpenMenuAriaLabel` | open menu | [components/dropdown/dropdown.templ:109](../components/dropdown/dropdown.templ#L109), [components/dropdown/dropdown.templ:134](../components/dropdown/dropdown.templ#L134) |
| `EmptyState.Title` | Nothing here yet | [components/emptystate/emptystate.templ:20](../components/emptystate/emptystate.templ#L20) |
| `EmptyState.Description` | Items will appear here when they are available. | [components/emptystate/emptystate.templ:23](../components/emptystate/emptystate.templ#L23) |
| `FileInput.BrowseLabel` | Browse | [components/fileinput/fileinput.templ:61](../components/fileinput/fileinput.templ#L61), [components/fileinput/fileinput.templ:104](../components/fileinput/fileinput.templ#L104) |
| `FileInput.DropText` |  or drag and drop here | [components/fileinput/fileinput.templ:63](../components/fileinput/fileinput.templ#L63) |
| `FileInput.EmptyText` | No file selected | [components/fileinput/fileinput.templ:102](../components/fileinput/fileinput.templ#L102) |
| `Form.EditLabel` | Edit | [components/form/expressions.go:11](../components/form/expressions.go#L11) |
| `Form.DoneLabel` | Done | [components/form/expressions.go:12](../components/form/expressions.go#L12) |
| `Form.ErrorTitle` | Validation failed | [components/form/expressions.go:18](../components/form/expressions.go#L18) |
| `Modal.CloseAriaLabel` | close modal | [components/modal/modal.templ:65](../components/modal/modal.templ#L65), [components/modal/modal.templ:125](../components/modal/modal.templ#L125) |
| `Modal.DialogCloseAriaLabel` | Close dialog | [components/modal/dialog.templ:29](../components/modal/dialog.templ#L29) |
| `Navbar.AriaLabel` | main navigation | [components/navbar/navbar.templ:41](../components/navbar/navbar.templ#L41) |
| `Navbar.OpenMenuAriaLabel` | Open mobile menu | [components/navbar/navbar.templ:85](../components/navbar/navbar.templ#L85) |
| `Navbar.CloseMenuAriaLabel` | Close mobile menu | [components/navbar/navbar.templ:85](../components/navbar/navbar.templ#L85) |
| `Navbar.UserMenuAriaLabel` | user menu | [components/navbar/navbar.templ:165](../components/navbar/navbar.templ#L165) |
| `Navbar.SecondaryAriaLabel` | secondary navigation | [components/navbar/expressions.go:11](../components/navbar/expressions.go#L11) |
| `Pagination.AriaLabel` | pagination | [components/pagination/pagination.templ:39](../components/pagination/pagination.templ#L39) |
| `Pagination.PreviousAriaLabel` | previous page | [components/pagination/pagination.templ:50](../components/pagination/pagination.templ#L50), [components/pagination/pagination.templ:62](../components/pagination/pagination.templ#L62) |
| `Pagination.NextAriaLabel` | next page | [components/pagination/pagination.templ:142](../components/pagination/pagination.templ#L142), [components/pagination/pagination.templ:154](../components/pagination/pagination.templ#L154) |
| `Pagination.MoreAriaLabel` | more pages | [components/pagination/pagination.templ:80](../components/pagination/pagination.templ#L80) |
| `Pagination.PreviousLabel` | Previous | [components/pagination/pagination.templ:56](../components/pagination/pagination.templ#L56), [components/pagination/pagination.templ:65](../components/pagination/pagination.templ#L65), [components/pagination/pagination.templ:71](../components/pagination/pagination.templ#L71) |
| `Pagination.NextLabel` | Next | [components/pagination/pagination.templ:147](../components/pagination/pagination.templ#L147), [components/pagination/pagination.templ:156](../components/pagination/pagination.templ#L156), [components/pagination/pagination.templ:162](../components/pagination/pagination.templ#L162) |
| `Pagination.PageAriaLabel` | Complete application-formatted message callback | [components/pagination/pagination.templ:91](../components/pagination/pagination.templ#L91), [components/pagination/pagination.templ:103](../components/pagination/pagination.templ#L103), [components/pagination/pagination.templ:115](../components/pagination/pagination.templ#L115), [components/pagination/pagination.templ:126](../components/pagination/pagination.templ#L126) |
| `Palette.Placeholder` | Pick a color | [components/palette/palette.templ:18](../components/palette/palette.templ#L18) |
| `Palette.ResetLabel` | Reset | [components/palette/palette.templ:24](../components/palette/palette.templ#L24) |
| `Palette.CustomLabel` | Custom color | [components/palette/palette.templ:49](../components/palette/palette.templ#L49), [components/palette/palette.templ:58](../components/palette/palette.templ#L58) |
| `Palette.WhiteLabel` | white | [components/palette/palette.templ:90](../components/palette/palette.templ#L90) |
| `Palette.BlackLabel` | black | [components/palette/palette.templ:101](../components/palette/palette.templ#L101) |
| `Popover.TriggerLabel` | Open | [components/popover/popover.templ:38](../components/popover/popover.templ#L38) |
| `Rating.Label` | Rating | [components/rating/rating.templ:27](../components/rating/rating.templ#L27), [components/rating/rating.templ:29](../components/rating/rating.templ#L29), [components/rating/rating.templ:48](../components/rating/rating.templ#L48) |
| `Rating.VeryDissatisfiedLabel` | very dissatisfied | [components/rating/expressions.go:11](../components/rating/expressions.go#L11) |
| `Rating.DissatisfiedLabel` | dissatisfied | [components/rating/expressions.go:11](../components/rating/expressions.go#L11) |
| `Rating.NeutralLabel` | neutral | [components/rating/expressions.go:11](../components/rating/expressions.go#L11) |
| `Rating.SatisfiedLabel` | satisfied | [components/rating/expressions.go:11](../components/rating/expressions.go#L11) |
| `Rating.VerySatisfiedLabel` | very satisfied | [components/rating/expressions.go:11](../components/rating/expressions.go#L11) |
| `Rating.StarLabel` | Complete application-formatted message callback | [components/rating/expressions.go:16](../components/rating/expressions.go#L16) |
| `SchemaForm.RequiredAriaLabel` | required | [components/schemaform/schemaform.templ:61](../components/schemaform/schemaform.templ#L61) |
| `SchemaForm.ManagedTitle` | Managed by platform — cannot override | [components/schemaform/schemaform.templ:219](../components/schemaform/schemaform.templ#L219) |
| `SchemaForm.ManagedLabel` | managed | [components/schemaform/schemaform.templ:223](../components/schemaform/schemaform.templ#L223) |
| `SchemaForm.ItemPlaceholder` | item | [components/schemaform/schemaform.templ:187](../components/schemaform/schemaform.templ#L187) |
| `SchemaForm.AddLabel` | + Add item | [components/schemaform/schemaform.templ:188](../components/schemaform/schemaform.templ#L188) |
| `SchemaForm.DefaultLabel` | Complete application-formatted message callback | [components/schemaform/schemaform.templ:150](../components/schemaform/schemaform.templ#L150) |
| `SchemaTree.RequiredLabel` | required | [components/schematree/schematree.templ:68](../components/schematree/schematree.templ#L68) |
| `SchemaTree.OptionalLabel` | optional | [components/schematree/schematree.templ:70](../components/schematree/schematree.templ#L70) |
| `SchemaTree.NullableLabel` | nullable | [components/schematree/schematree.templ:73](../components/schematree/schematree.templ#L73) |
| `SchemaTree.DeprecatedLabel` | deprecated | [components/schematree/schematree.templ:76](../components/schematree/schematree.templ#L76) |
| `Search.Label` | Search | [components/search/expressions.go:11](../components/search/expressions.go#L11) |
| `Search.Placeholder` | Search | [components/search/expressions.go:12](../components/search/expressions.go#L12) |
| `Search.ShortcutText` | ⌘ K | [components/search/expressions.go:13](../components/search/expressions.go#L13) |
| `Search.EscapeText` | Esc | [components/search/expressions.go:14](../components/search/expressions.go#L14) |
| `Search.EmptyText` | No results found. | [components/search/expressions.go:15](../components/search/expressions.go#L15) |
| `Search.ResultsAriaLabel` | Complete application-formatted message callback | [components/search/search.templ:125](../components/search/search.templ#L125) |
| `Select.Placeholder` | Please Select | [components/select/expressions.go:11](../components/select/expressions.go#L11) |
| `Select.ListAriaLabel` | Complete application-formatted message callback | [components/select/select.templ:127](../components/select/select.templ#L127), [components/select/select.templ:140](../components/select/select.templ#L140) |
| `Sidebar.AriaLabel` | sidebar navigation | [components/sidebar/sidebar.templ:22](../components/sidebar/sidebar.templ#L22) |
| `Sidebar.SearchAriaLabel` | Search | [components/sidebar/sidebar.templ:53](../components/sidebar/sidebar.templ#L53) |
| `Sidebar.SkipLabel` | skip to the main content | [components/sidebar/sidebar.templ:26](../components/sidebar/sidebar.templ#L26) |
| `Sidebar.ActiveLabel` | active | [components/sidebar/sidebar.templ:175](../components/sidebar/sidebar.templ#L175), [components/sidebar/sidebar.templ:252](../components/sidebar/sidebar.templ#L252) |
| `Sidebar.TriggerLabel` | Open sidebar | [components/sidebar/expressions.go:12](../components/sidebar/expressions.go#L12) |
| `Skeleton.AriaLabel` | Loading content | [components/skeleton/expressions.go:12](../components/skeleton/expressions.go#L12) |
| `SplitButton.MenuAriaLabel` | More actions | [components/splitbutton/expressions.go:12](../components/splitbutton/expressions.go#L12) |
| `Steps.AriaLabel` | progress | [components/steps/expressions.go:11](../components/steps/expressions.go#L11) |
| `Steps.CompletedLabel` | completed | [components/steps/steps.templ:101](../components/steps/steps.templ#L101) |
| `StructuredInput.AddLabel` | Add row | [components/structuredinput/expressions.go:11](../components/structuredinput/expressions.go#L11) |
| `StructuredInput.RemoveAriaLabel` | Remove row | [components/structuredinput/structuredinput.templ:45](../components/structuredinput/structuredinput.templ#L45) |
| `Table.LoadingText` | Loading... | [components/table/table.templ:264](../components/table/table.templ#L264), [components/table/table.templ:746](../components/table/table.templ#L746) |
| `Table.ActionsLabel` | Actions | [components/table/table.templ:315](../components/table/table.templ#L315) |
| `Table.LoadMoreLabel` | Load more | [components/table/table.templ:696](../components/table/table.templ#L696) |
| `Table.FiltersLabel` | Filters | [components/table/table.templ:173](../components/table/table.templ#L173) |
| `Table.OpenRowLabel` | Complete application-formatted message callback | [components/table/table.templ:439](../components/table/table.templ#L439), [components/table/table.templ:441](../components/table/table.templ#L441) |
| `Tabs.OptionsAriaLabel` | tab options | [components/tabs/tabs.templ:33](../components/tabs/tabs.templ#L33) |
| `Tabs.LoadingText` | Loading... | [components/tabs/tabs.templ:140](../components/tabs/tabs.templ#L140) |
| `TagsList.AddLabel` | Add | [components/tagslist/expressions.go:11](../components/tagslist/expressions.go#L11) |
| `TagsList.Placeholder` | Add a tag... | [components/tagslist/expressions.go:12](../components/tagslist/expressions.go#L12) |
| `TagsList.RemoveAriaLabel` | Remove tag | [components/tagslist/tagslist.templ:35](../components/tagslist/tagslist.templ#L35) |
| `TextInput.ShowPasswordAriaLabel` | Show password | [components/textinput/textinput.templ:154](../components/textinput/textinput.templ#L154) |
| `TextInput.SearchAriaLabel` | search | [components/textinput/textinput.templ:211](../components/textinput/textinput.templ#L211) |
| `Textarea.SendAriaLabel` | send | [components/textarea/textarea.templ:98](../components/textarea/textarea.templ#L98) |
| `Textarea.EmojiAriaLabel` | Emojis | [components/textarea/textarea.templ:90](../components/textarea/textarea.templ#L90) |
| `Textarea.AttachAriaLabel` | Attach a file | [components/textarea/textarea.templ:91](../components/textarea/textarea.templ#L91) |
| `Textarea.VoiceAriaLabel` | Send voice | [components/textarea/textarea.templ:92](../components/textarea/textarea.templ#L92) |
| `Textarea.SendLabel` | Send | [components/textarea/textarea.templ:100](../components/textarea/textarea.templ#L100) |
| `Toast.DismissLabel` | Dismiss | [components/toast/expressions.go:11](../components/toast/expressions.go#L11), [components/toast/toast.templ:166](../components/toast/toast.templ#L166) |
| `Toast.DismissAriaLabel` | dismiss notification | [components/toast/toast.templ:128](../components/toast/toast.templ#L128), [components/toast/toast.templ:171](../components/toast/toast.templ#L171), [components/toast/toast.templ:296](../components/toast/toast.templ#L296), [components/toast/toast.templ:361](../components/toast/toast.templ#L361) |
| `Toolbar.AriaLabel` | Page tools | [components/toolbar/expressions.go:12](../components/toolbar/expressions.go#L12) |
| `Tooltip.TriggerLabel` | Hover Me | [components/tooltip/tooltip.templ:61](../components/tooltip/tooltip.templ#L61), [components/tooltip/tooltip.templ:85](../components/tooltip/tooltip.templ#L85), [components/tooltip/tooltip.templ:113](../components/tooltip/tooltip.templ#L113) |

Browser sources: `assets/js/src/components/code-block.js` consumes copy state text;
`combobox-client.js` consumes selected-count labels; `select.js` consumes its resolved
placeholder. Carousel, Navbar, Palette, and FileInput bind escaped data attributes.

Intentionally retained: caller content, diagnostic errors, technical/protocol values,
decorative image alternative text, punctuation-only separators, badge count capping,
and raw numeric display/input formatting. Runtime English literals remain fallbacks
for older markup lacking new data attributes. Some private helpers retain English
fallbacks for unconfigured values; rendering supplies resolved values before calling them.

See [the expression guide](EXPRESSIONS.md) for precedence, request handling, concurrency,
dynamic messages, and the application/library boundary.
