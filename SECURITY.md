# Security Policy

Goshtoso is alpha-stage software. We take security seriously and appreciate
responsible disclosure.

## Supported versions

| Version | Supported |
|---------|-----------|
| `v0.0.x` (alpha) | ✅ latest tag only |
| older | ❌ |

During alpha, only the most recent tagged release receives fixes.

## Reporting a vulnerability

**Do not open a public issue for security problems.**

Report privately via GitHub's
[**Report a vulnerability**](https://github.com/araihu/goshtoso/security/advisories/new)
(Security → Advisories), or email **security@araihu.dev**.

Please include:

- affected component / endpoint and version (tag or commit),
- a description and impact assessment,
- reproduction steps or a proof of concept,
- any suggested remediation.

## What to expect

- **Acknowledgement** within 5 business days.
- An initial assessment and severity rating shortly after.
- Coordinated disclosure: we'll agree on a timeline and credit you (unless you
  prefer to remain anonymous) once a fix is released.

## Scope notes

Goshtoso renders HTML via templ and wires Alpine.js / HTMX. Of particular
interest: HTML/attribute injection through component config, unsafe
`templ.Raw` usage, and the table/example HTTP endpoints
(`/api/components/...`, `/api/examples/...`). The demo server is for local
development and is not hardened for untrusted public deployment.
