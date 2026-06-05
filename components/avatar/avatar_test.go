package avatar

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestAvatarBorderClassesRenderOnRoot(t *testing.T) {
	var buf bytes.Buffer
	err := Avatar(Config{
		Src:         "/avatar.webp",
		Alt:         "Bordered avatar",
		Border:      true,
		BorderColor: "border-success",
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render avatar: %v", err)
	}

	html := buf.String()
	for _, want := range []string{"border-2", "border-success", "p-0.5"} {
		if !bytes.Contains([]byte(html), []byte(want)) {
			t.Fatalf("bordered avatar missing %q in rendered HTML:\n%s", want, html)
		}
	}
	if bytes.Contains([]byte(html), []byte("border-outline")) {
		t.Fatalf("bordered avatar root should not also render outline border classes:\n%s", html)
	}
}

func TestAvatarBorderLayersRespectPadding(t *testing.T) {
	var buf bytes.Buffer
	err := Avatar(Config{
		Src:         "/avatar.webp",
		Alt:         "Bordered avatar",
		Border:      true,
		BorderColor: "border-success",
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render avatar: %v", err)
	}

	html := buf.String()
	for _, want := range []string{"p-0.5", "relative block size-full overflow-hidden rounded-full"} {
		if !bytes.Contains([]byte(html), []byte(want)) {
			t.Fatalf("bordered avatar layers should render inside a clipped content box; missing %q:\n%s", want, html)
		}
	}
}

func TestAvatarSquareRadiusClassesRenderOnRoot(t *testing.T) {
	var buf bytes.Buffer
	err := Avatar(Config{
		Initials: "SQ",
		Shape:    ShapeSquare,
		Radius:   RadiusLG,
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render avatar: %v", err)
	}

	html := buf.String()
	if !bytes.Contains([]byte(html), []byte("rounded-lg")) {
		t.Fatalf("square avatar missing rounded-lg in rendered HTML:\n%s", html)
	}
	if bytes.Contains([]byte(html), []byte("rounded-full")) {
		t.Fatalf("square avatar should not render rounded-full:\n%s", html)
	}
}

func TestAvatarReactiveRadiusBindsSizeAndRadiusClasses(t *testing.T) {
	var buf bytes.Buffer
	err := Avatar(Config{
		Initials:       "SQ",
		Shape:          ShapeSquare,
		Reactive:       true,
		ReactiveRadius: true,
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render avatar: %v", err)
	}

	html := buf.String()
	if !bytes.Contains([]byte(html), []byte(`x-bind:class="[avatarSizeClass, avatarRadiusClass]"`)) {
		t.Fatalf("reactive square avatar missing combined size/radius binding:\n%s", html)
	}
	for _, want := range []string{"relative block size-full overflow-hidden", `x-bind:class="avatarRadiusClass"`} {
		if !bytes.Contains([]byte(html), []byte(want)) {
			t.Fatalf("reactive square avatar inner clip box should bind avatarRadiusClass; missing %q:\n%s", want, html)
		}
	}
}

func TestAvatarStackRendersOverlappingItems(t *testing.T) {
	var buf bytes.Buffer
	err := AvatarStack(StackConfig{
		Label: "Project members",
		Items: []Config{
			{Src: "/a.webp", Alt: "Ada Lovelace"},
			{Src: "/g.webp", Alt: "Grace Hopper"},
			{Initials: "KT", Variant: Primary},
		},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render avatar stack: %v", err)
	}

	html := buf.String()
	for _, want := range []string{"aria-label=\"Project members\"", "-ml-3", "ring-2", "Ada Lovelace", "Grace Hopper", "KT"} {
		if !bytes.Contains([]byte(html), []byte(want)) {
			t.Fatalf("avatar stack missing %q in rendered HTML:\n%s", want, html)
		}
	}
}

func TestAvatarEscapesSrcInAlpineErrorHandler(t *testing.T) {
	var buf bytes.Buffer
	err := Avatar(Config{
		Src: "data:image/svg+xml,<svg viewBox='0 0 1 1'>\n\r\t\u2028\u2029",
		Alt: "Injected avatar",
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render avatar: %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `console.warn(&#39;[avatar] failed to load image:&#39;, &#39;data:image/svg+xml,&lt;svg viewBox=&#39;0 0 1 1&#39;&gt;`) {
		t.Fatalf("avatar rendered raw Src inside Alpine error handler:\n%s", html)
	}
	for _, want := range []string{`viewBox=\&#39;0 0 1 1\&#39;`, `\n\r\t\u2028\u2029`} {
		if !strings.Contains(html, want) {
			t.Fatalf("avatar missing escaped Src fragment %q in rendered HTML:\n%s", want, html)
		}
	}
}
