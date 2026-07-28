# Application-patterns component-shell snag journal

Date: 2026-07-27
Worktree: `/private/tmp/gs-visual-defects`

## TOC test retained the pre-shell header contract

The application-patterns E2E test checked anchor visibility against
`header[data-boot-anim="header"]`, a selector from the retired in-repo layout.
The component-doc-shell migration renders `.component-doc-shell__header`
instead, so the assertion was permanently false despite correct TOC scrolling.

Inspecting the component-doc-shell runtime also confirmed that TOC links scroll
smoothly inside `#page-scroll`; the test now waits for the settled position
against the shell's public header class rather than measuring immediately after
the click.
