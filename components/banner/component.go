package banner

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable banner component.
type Instance struct {
	expressionOverrides expressions.Banner
	cfg                 Config
}

// Banner returns a renderable banner component.
func Banner(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a banner.
func (Instance) Kind() components.Kind {
	return components.KindBanner
}

// Render writes the banner markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Banner: i.expressionOverrides})
	return bannerTemplate(i.cfg).Render(ctx, w)
}

// CookieBannerInstance is a renderable cookie consent banner.
type CookieBannerInstance struct {
	expressionOverrides expressions.Banner
	cfg                 CookieBannerConfig
}

// CookieBanner returns a renderable cookie consent banner.
func CookieBanner(cfg CookieBannerConfig) CookieBannerInstance {
	return CookieBannerInstance{cfg: cfg}
}

// Kind identifies the component as a cookie consent banner.
func (CookieBannerInstance) Kind() components.Kind {
	return components.KindCookieBanner
}

// Render writes the cookie consent banner markup.
func (i CookieBannerInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Banner: i.expressionOverrides})
	return cookieBannerTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = CookieBannerInstance{}
)

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Banner) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Banner: i.expressionOverrides}, expressions.Set{Banner: values}).Banner
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i CookieBannerInstance) WithExpressions(values expressions.Banner) CookieBannerInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Banner: i.expressionOverrides}, expressions.Set{Banner: values}).Banner
	return i
}
