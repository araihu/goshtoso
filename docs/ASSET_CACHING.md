# Asset caching

Goshtoso classifies browser caching from the identity carried by an asset URL.
Use `assets.CacheControl(path)` when serving exported files, or wrap an HTTP
handler with `assets.WithCacheControl(handler)`.

Two policies exist:

- Exact semantic-version runtime and license paths, numeric Charts control
  generations, and content-hash path segments use
  `public, max-age=31536000, immutable`.
- Unversioned CSS, JavaScript, icons, logos, images, fallbacks, and root aliases
  use `public, max-age=0, must-revalidate`.

Examples:

```text
/assets/js/runtime/alpinejs/3.14.9/alpine.min.js  immutable
/charts/assets/js/controls/5/controls.js          immutable
/assets/app.3c91e915370d.js                       immutable
/assets/styles.css                                revalidate
/assets/js/goshtoso.min.js                        revalidate
/assets/icons/heroicons.svg                       revalidate
/assets/images/goshtoso-logo.svg                  revalidate
/favicon.svg                                      revalidate
```

Query parameters do not prove asset identity. A URL such as
`/assets/styles.css?v=3c91e915370d` remains a mutable alias because the path can
still return different bytes. Put the version or content hash in the path and
make the handler reject unknown identities before relying on immutable caching.

`assets.Handler()` applies this policy. The demo server uses the same wrapper
for library assets, site-only JavaScript, component-doc shell assets, Charts
assets, and root favicon or manifest aliases. Exporters using `assets.FS()` or
`assets.ReadFile()` own their HTTP route and should apply the same classifier.
