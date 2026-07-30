package micromessage

import (
	"math"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
)

func RGB(r, g, b float64) *colorful.Color {
	return &colorful.Color{R: r, G: g, B: b}
}

func FromHex(s string) *colorful.Color {
	col, _ := colorful.Hex(s)
	return &col
}

type namedColor struct {
	name string
	c    colorful.Color
}

func hex(hex string) colorful.Color {
	col, _ := colorful.Hex(hex)
	return col
}

var namedColors = []namedColor{
	{"black", hex("#000000")},
	{"dark_blue", hex("#0000AA")},
	{"dark_green", hex("#00AA00")},
	{"dark_aqua", hex("#00AAAA")},
	{"dark_red", hex("#AA0000")},
	{"dark_purple", hex("#AA00AA")},
	{"gold", hex("#FFAA00")},
	{"gray", hex("#AAAAAA")},
	{"dark_gray", hex("#555555")},
	{"blue", hex("#5555FF")},
	{"green", hex("#55FF55")},
	{"aqua", hex("#55FFFF")},
	{"red", hex("#FF5555")},
	{"light_purple", hex("#FF55FF")},
	{"yellow", hex("#FFFF55")},
	{"white", hex("#FFFFFF")},
}

var namedColorAliases = map[string]string{
	"grey":      "gray",
	"dark_grey": "dark_gray",
}

var namedColorByName = func() map[string]colorful.Color {
	m := make(map[string]colorful.Color, len(namedColors))
	for _, nc := range namedColors {
		m[nc.name] = nc.c
	}
	return m
}()

var namedColorByRGB = func() map[[3]float64]string {
	m := make(map[[3]float64]string, len(namedColors))
	for _, nc := range namedColors {
		m[[3]float64{nc.c.R, nc.c.G, nc.c.B}] = nc.name
	}
	return m
}()

func ResolveColor(name string) *colorful.Color {
	name = strings.ToLower(name)
	if canon, ok := namedColorAliases[name]; ok {
		name = canon
	}
	if c, ok := namedColorByName[name]; ok {
		cc := c
		return &cc
	}
	if strings.HasPrefix(name, "#") {
		return FromHex(name)
	}
	return nil
}

// Accepts a Color and outputs a named color if matching
func NamedColorName(c *colorful.Color) string {
	if c == nil {
		return ""
	}
	if name, ok := namedColorByRGB[[3]float64{c.R, c.G, c.B}]; ok {
		return name
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
