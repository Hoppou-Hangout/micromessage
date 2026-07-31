package micromessage

import (
	"math"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	mcolor "go.minekube.com/common/minecraft/color"
)

func RGB(r, g, b float64) *colorful.Color {
	return &colorful.Color{R: r, G: g, B: b}
}

func FromHex(s string) *colorful.Color {
	col, _ := colorful.Hex(s)
	return &col
}

// Aliases minekube's color.Names doesn't cover (British spellings).
var namedColorAliases = map[string]string{
	"grey":      "gray",
	"dark_grey": "dark_gray",
}

// ResolveColor resolves a named Minecraft color (delegating to minekube's
// color.Names) or a #rrggbb / #rgb hex literal into a *colorful.Color.
// colorful is kept as the public exchange type since minekube's color.RGB
// is a colorful.Color underneath.
func ResolveColor(name string) *colorful.Color {
	name = strings.ToLower(name)
	if canon, ok := namedColorAliases[name]; ok {
		name = canon
	}
	if n, ok := mcolor.Names[name]; ok {
		cc := colorful.Color(*n.RGB)
		return &cc
	}
	if strings.HasPrefix(name, "#") {
		return FromHex(name)
	}
	return nil
}

// NamedColorName returns the Minecraft name of c if it exactly matches one of
// minekube's named colors, or "" otherwise. (We do not use NearestNamed because
// that would report a nearest match for arbitrary colors.)
func NamedColorName(c *colorful.Color) string {
	if c == nil {
		return ""
	}
	rgb := (*mcolor.RGB)(c)
	for _, n := range mcolor.NamesOrder {
		if *n.RGB == *rgb {
			return n.Name
		}
	}
	return ""
}

func LerpColor(t float64, a, b *colorful.Color) *colorful.Color {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	lerp := func(x, y float64) float64 {
		return x + t*(y-x)
	}
	return &colorful.Color{R: lerp(a.R, b.R), G: lerp(a.G, b.G), B: lerp(a.B, b.B)}
}

func HSVColor(h, s, v float64) *colorful.Color {
	i := math.Floor(h * 6)
	f := h*6 - i
	p := v * (1 - s)
	q := v * (1 - f*s)
	t := v * (1 - (1-f)*s)

	var r, g, b float64
	switch int(i) % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	case 5:
		r, g, b = v, p, q
	}

	return &colorful.Color{R: r, G: g, B: b}
}