package micromessage

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Color struct {
	R, G, B uint8
}

func RGB(r, g, b uint8) *Color {
	return &Color{R: r, G: g, B: b}
}

func FromHex(s string) *Color {
	if len(s) != 7 || s[0] != '#' {
		return nil
	}
	v, err := strconv.ParseInt(s[1:], 16, 32)
	if err != nil {
		return nil
	}
	return &Color{R: uint8((v >> 16) & 0xFF), G: uint8((v >> 8) & 0xFF), B: uint8(v & 0xFF)}
}

func (c *Color) HexString() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func (c *Color) Equal(o *Color) bool {
	if c == nil || o == nil {
		return c == o
	}
	return c.R == o.R && c.G == o.G && c.B == o.B
}

type namedColor struct {
	name string
	c    Color
}

var namedColors = []namedColor{
	{"black", Color{0x00, 0x00, 0x00}},
	{"dark_blue", Color{0x00, 0x00, 0xAA}},
	{"dark_green", Color{0x00, 0xAA, 0x00}},
	{"dark_aqua", Color{0x00, 0xAA, 0xAA}},
	{"dark_red", Color{0xAA, 0x00, 0x00}},
	{"dark_purple", Color{0xAA, 0x00, 0xAA}},
	{"gold", Color{0xFF, 0xAA, 0x00}},
	{"gray", Color{0xAA, 0xAA, 0xAA}},
	{"dark_gray", Color{0x55, 0x55, 0x55}},
	{"blue", Color{0x55, 0x55, 0xFF}},
	{"green", Color{0x55, 0xFF, 0x55}},
	{"aqua", Color{0x55, 0xFF, 0xFF}},
	{"red", Color{0xFF, 0x55, 0x55}},
	{"light_purple", Color{0xFF, 0x55, 0xFF}},
	{"yellow", Color{0xFF, 0xFF, 0x55}},
	{"white", Color{0xFF, 0xFF, 0xFF}},
}

var namedColorAliases = map[string]string{
	"grey":      "gray",
	"dark_grey": "dark_gray",
}

var namedColorByName = func() map[string]Color {
	m := make(map[string]Color, len(namedColors))
	for _, nc := range namedColors {
		m[nc.name] = nc.c
	}
	return m
}()

var namedColorByRGB = func() map[[3]uint8]string {
	m := make(map[[3]uint8]string, len(namedColors))
	for _, nc := range namedColors {
		m[[3]uint8{nc.c.R, nc.c.G, nc.c.B}] = nc.name
	}
	return m
}()

func ResolveColor(name string) *Color {
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
func NamedColorName(c *Color) string {
	if c == nil {
		return ""
	}
	if name, ok := namedColorByRGB[[3]uint8{c.R, c.G, c.B}]; ok {
		return name
	}
	return ""
}

func LerpColor(t float64, a, b *Color) *Color {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	lerp := func(x, y uint8) uint8 {
		return uint8(math.Round(float64(x) + t*(float64(y)-float64(x))))
	}
	return &Color{R: lerp(a.R, b.R), G: lerp(a.G, b.G), B: lerp(a.B, b.B)}
}

func HSVColor(h, s, v float64) *Color {
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
	clamp := func(x float64) uint8 {
		return uint8(math.Round(x * 255))
	}
	return &Color{R: clamp(r), G: clamp(g), B: clamp(b)}
}
