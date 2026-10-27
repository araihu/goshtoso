package themes

import (
	"fmt"
	"maps"
	"math"
	"strconv"
	"strings"
)

// CSS returns a literal CSS color. Renderers must Validate the definition first.
func (c Color) CSS() string {
	if c.Space == "oklch" {
		return fmt.Sprintf("oklch(%g %g %g / %g)", c.Channels[0], c.Channels[1], c.Channels[2], c.Alpha)
	}
	return fmt.Sprintf("color(srgb %g %g %g / %g)", c.Channels[0], c.Channels[1], c.Channels[2], c.Alpha)
}
func clamp(v float64) float64 { return math.Max(0, math.Min(1, v)) }
func (c Color) rgb() [3]float64 {
	if c.Space == "srgb" {
		return c.Channels
	}
	l, ch, h := c.Channels[0], c.Channels[1], c.Channels[2]*math.Pi/180
	a, b := ch*math.Cos(h), ch*math.Sin(h)
	ll, mm, ss := l+.3963377774*a+.2158037573*b, l-.1055613458*a-.0638541728*b, l-.0894841775*a-1.291485548*b
	ll, mm, ss = ll*ll*ll, mm*mm*mm, ss*ss*ss
	result := [3]float64{4.0767416621*ll - 3.3077115913*mm + .2309699292*ss, -1.2684380046*ll + 2.6097574011*mm - .3413193965*ss, -.0041960863*ll - .7034186147*mm + 1.707614701*ss}
	for i, v := range result {
		if v <= .0031308 {
			result[i] = 12.92 * v
		} else {
			result[i] = 1.055*math.Pow(v, 1/2.4) - .055
		}
	}
	return result
}
func mix(a, b Color, ratio float64) Color {
	ar, br := a.rgb(), b.rgb()
	alpha := a.Alpha*ratio + b.Alpha*(1-ratio)
	c := Color{Space: "srgb", Alpha: alpha}
	for i := range c.Channels {
		if alpha > 0 {
			c.Channels[i] = clamp((ar[i]*a.Alpha*ratio + br[i]*b.Alpha*(1-ratio)) / alpha)
		}
	}
	return c
}
func splitArguments(s string) []string {
	out := []string{}
	depth, start := 0, 0
	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	return append(out, strings.TrimSpace(s[start:]))
}
func number(s string) (float64, error) {
	if before, ok := strings.CutSuffix(s, "%"); ok {
		v, e := strconv.ParseFloat(before, 64)
		return v / 100, e
	}
	return strconv.ParseFloat(s, 64)
}
func parseFunctional(s, prefix string) (Color, error) {
	fields := strings.Fields(strings.TrimSuffix(strings.TrimPrefix(s, prefix), ")"))
	c := Color{Space: "oklch", Alpha: 1}
	if prefix == "color(srgb " {
		c.Space = "srgb"
	}
	if len(fields) != 3 && len(fields) != 5 {
		return c, ErrInvalid
	}
	for i := range 3 {
		if i == 2 && c.Space == "oklch" && fields[i] == "none" && c.Channels[1] == 0 {
			continue
		}
		v, e := number(fields[i])
		if e != nil {
			return c, e
		}
		c.Channels[i] = v
	}
	if len(fields) == 5 {
		if fields[3] != "/" {
			return c, ErrInvalid
		}
		v, e := number(fields[4])
		if e != nil {
			return c, e
		}
		c.Alpha = v
	}
	return c, nil
}
func parseHex(s string) (Color, error) {
	h := strings.TrimPrefix(s, "#")
	if len(h) == 3 || len(h) == 4 {
		var b strings.Builder
		for _, r := range h {
			b.WriteRune(r)
			b.WriteRune(r)
		}
		h = b.String()
	}
	if len(h) != 6 && len(h) != 8 {
		return Color{}, ErrInvalid
	}
	v, e := strconv.ParseUint(h, 16, 32)
	if e != nil {
		return Color{}, e
	}
	alpha := 1.0
	if len(h) == 8 {
		alpha = float64(v&255) / 255
		v >>= 8
	}
	return Color{Space: "srgb", Channels: [3]float64{float64(v>>16&255) / 255, float64(v>>8&255) / 255, float64(v&255) / 255}, Alpha: alpha}, nil
}

// ParseColor accepts a literal hex, sRGB, OKLCH, or transparent value.
func ParseColor(s string) (Color, error) {
	s = strings.TrimSpace(s)
	var c Color
	var err error
	switch {
	case s == "transparent":
		c = Color{Space: "srgb"}
	case s == "white":
		c = Color{Space: "srgb", Channels: [3]float64{1, 1, 1}, Alpha: 1}
	case s == "black":
		c = Color{Space: "srgb", Alpha: 1}
	case strings.HasPrefix(s, "#"):
		c, err = parseHex(s)
	case strings.HasPrefix(s, "oklch("):
		c, err = parseFunctional(s, "oklch(")
	case strings.HasPrefix(s, "color(srgb "):
		c, err = parseFunctional(s, "color(srgb ")
	default:
		err = ErrInvalid
	}
	if err != nil || !c.Valid() {
		return Color{}, ErrInvalid
	}
	return c, nil
}
func resolveColor(expr string, vars map[string]string, seen map[string]bool) (Color, error) {
	expr = strings.TrimSpace(expr)
	if after, ok := strings.CutPrefix(expr, "var("); ok {
		key := strings.TrimSuffix(after, ")")
		if seen[key] {
			return Color{}, ErrInvalid
		}
		value, ok := vars[key]
		if !ok {
			return Color{}, fmt.Errorf("missing token %s", key)
		}
		next := map[string]bool{}
		maps.Copy(next, seen)
		next[key] = true
		return resolveColor(value, vars, next)
	}
	if after, ok := strings.CutPrefix(expr, "color-mix("); ok {
		parts := splitArguments(strings.TrimSuffix(after, ")"))
		if len(parts) != 3 || parts[0] != "in srgb" {
			return Color{}, ErrInvalid
		}
		i := strings.LastIndex(parts[1], " ")
		if i < 0 {
			return Color{}, ErrInvalid
		}
		ratio, e := number(parts[1][i+1:])
		if e != nil {
			return Color{}, e
		}
		j := strings.LastIndex(parts[2], " ")
		if j < 0 {
			return Color{}, ErrInvalid
		}
		a, e := resolveColor(parts[1][:i], vars, seen)
		if e != nil {
			return Color{}, e
		}
		b, e := resolveColor(parts[2][:j], vars, seen)
		if e != nil {
			return Color{}, e
		}
		return mix(a, b, ratio), nil
	}
	c, err := ParseColor(expr)
	if err != nil {
		return Color{}, fmt.Errorf("literal %q: %w", expr, err)
	}
	return c, nil
}

// Contrast computes an advisory contrast ratio after conversion to sRGB.
func Contrast(a, b Color) float64 {
	lum := func(c Color) float64 {
		rgb := c.rgb()
		for i, v := range rgb {
			v = clamp(v)
			if v <= .04045 {
				rgb[i] = v / 12.92
			} else {
				rgb[i] = math.Pow((v+.055)/1.055, 2.4)
			}
		}
		return .2126*rgb[0] + .7152*rgb[1] + .0722*rgb[2]
	}
	l1, l2 := lum(a), lum(b)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + .05) / (l2 + .05)
}
