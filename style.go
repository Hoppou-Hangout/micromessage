package micromessage

import (
	"math"
	"strconv"
	"strings"

	mccolor "go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
)

// resolveColor turns a color name or "#rrggbb"/"#rgb" hex literal into a
// mccolor.Color, using minekube/common's own named-color table and hex parser
// so results are identical to what the rest of a Gate-based stack expects.
func resolveColor(name string) (mccolor.Color, bool) {
	if strings.HasPrefix(name, "#") {
		if rgb, err := mccolor.Hex(name); err == nil {
			return rgb, true
		}
		return nil, false
	}
	lower := strings.ToLower(name)
	// A couple of common British-spelling / legacy aliases that
	// minekube/common's Names map doesn't itself carry.
	switch lower {
	case "grey":
		lower = "gray"
	case "dark_grey":
		lower = "dark_gray"
	}
	if named, ok := mccolor.Names[lower]; ok {
		return named, true
	}
	return nil, false
}

// decorationTags maps every accepted spelling of a decoration to the
// component.Decoration it controls.
var decorationTags = map[string]c.Decoration{
	"bold": c.Bold, "b": c.Bold,
	"italic": c.Italic, "i": c.Italic, "em": c.Italic,
	"underlined": c.Underlined, "u": c.Underlined,
	"strikethrough": c.Strikethrough, "st": c.Strikethrough,
	"obfuscated": c.Obfuscated, "obf": c.Obfuscated,
}

// clickActionAliases covers the few click actions MiniMessage names differently
// than component.ClickActions' canonical keys, if any turn up.
func clickAction(name string) (c.ClickAction, bool) {
	a, ok := c.ClickActions[strings.ToLower(name)]
	return a, ok
}

// applyTag returns the style that should apply to this tag's own content, given
// the style inherited from its parent. "gradient" and "reset" are handled
// structurally by the caller, not here.
func applyTag(style c.Style, node *Node) c.Style {
	name := node.Name
	negate := false
	if strings.HasPrefix(name, "!") {
		negate = true
		name = name[1:]
	}
	lower := strings.ToLower(name)

	if dec, ok := decorationTags[lower]; ok {
		value := !negate
		if len(node.Args) == 1 {
			switch strings.ToLower(node.Args[0]) {
			case "true":
				value = true
			case "false":
				value = false
			}
		}
		style.SetDecoration(dec, c.StateByBool(value))
		return style
	}

	switch lower {
	case "color", "colour", "c":
		if len(node.Args) > 0 {
			if col, ok := resolveColor(node.Args[0]); ok {
				style.Color = col
			}
		}
		return style

	case "click":
		if len(node.Args) >= 1 {
			if action, ok := clickAction(node.Args[0]); ok {
				value := ""
				if len(node.Args) > 1 {
					value = strings.Join(node.Args[1:], ":")
				}
				style.ClickEvent = c.NewClickEvent(action, value)
			}
		}
		return style

	case "hover":
		if len(node.Args) >= 1 && strings.EqualFold(node.Args[0], "show_text") {
			value := ""
			if len(node.Args) > 1 {
				value = strings.Join(node.Args[1:], ":")
			}
			// NOTE: real MiniMessage recursively parses this value as its own
			// MiniMessage content (so hover text can carry its own
			// colors/formatting). We only support a flat-text hover here; see
			// the package doc for details.
			style.HoverEvent = c.ShowText(&c.Text{Content: value})
		}
		// show_item / show_entity are not implemented.
		return style
	}

	// Bare color tags, e.g. <red> or <#ff0000>.
	if col, ok := resolveColor(name); ok {
		style.Color = col
	}
	return style
}

// --- Rendering ----------------------------------------------------------

// render walks a list of sibling nodes left to right, threading a mutable
// "current style" so <reset> (always self-closed, never wraps anything) can
// change the ambient style for the *remaining* siblings without touching ones
// already emitted.
func render(nodes []*Node, inherited c.Style) []c.Component {
	var out []c.Component
	cur := inherited

	for _, n := range nodes {
		switch n.Kind {
		case KindText:
			out = append(out, &c.Text{Content: n.Text, S: cur})

		case KindElement:
			lower := strings.ToLower(n.Name)
			switch lower {
			case "reset":
				cur = c.Style{}
			case "br", "newline":
				out = append(out, &c.Text{Content: "\n", S: cur})
			case "gradient":
				out = append(out, renderGradient(n, cur)...)
			default:
				childStyle := applyTag(cur, n)
				out = append(out, render(n.Children, childStyle)...)
			}
		}
	}
	return out
}

// --- Gradients ------------------------------------------------------------
//
// <gradient:c1:c2:...:cN[:phase]>text</gradient>
//
// This is a faithful port of Adventure's real GradientTag algorithm: colors
// interpolate across the flattened character span in "color-index space" with
// modulo-wrapped stops (so multi-color gradients cycle correctly), and a phase
// that for negative values reverses the color list and remaps into [0,1) before
// scaling into index space. No colors given defaults to white -> black,
// matching the real tag.
func renderGradient(node *Node, inherited c.Style) []c.Component {
	colors, phase := parseGradientArgs(node.Args)
	if len(colors) < 2 {
		colors = []rgb{hexRGB(0xffffff), hexRGB(0x000000)}
	}

	// Reset the ambient *color* here (but keep decorations/click/hover) so a
	// color inherited from outside the gradient doesn't get mistaken for an
	// explicit override inside it - only a color tag actually nested inside
	// this gradient should block the gradient from applying.
	base := inherited
	base.Color = nil
	runes, styles := flattenChars(node.Children, base)

	calc := newGradientCalc(colors, phase, len(runes))

	var out []c.Component
	var curText strings.Builder
	var curStyle *c.Style

	flush := func() {
		if curStyle != nil && curText.Len() > 0 {
			out = append(out, &c.Text{Content: curText.String(), S: *curStyle})
		}
		curText.Reset()
		curStyle = nil
	}

	for i, r := range runes {
		s := styles[i]
		if s.Color == nil {
			s.Color = rgbToColor(calc.at(i))
		}
		if curStyle == nil || !sameStyle(*curStyle, s) {
			flush()
			cp := s
			curStyle = &cp
		}
		curText.WriteRune(r)
	}
	flush()
	return out
}

// flattenChars walks nodes, resolving nested tags normally (including "reset"
// and nested "gradient"), and returns every rune of the eventual text alongside
// the style that applies to it, in order.
func flattenChars(nodes []*Node, inherited c.Style) ([]rune, []c.Style) {
	var runes []rune
	var styles []c.Style

	var walk func(nodes []*Node, style c.Style)
	walk = func(nodes []*Node, style c.Style) {
		local := style
		for _, n := range nodes {
			switch n.Kind {
			case KindText:
				for _, r := range n.Text {
					runes = append(runes, r)
					styles = append(styles, local)
				}
			case KindElement:
				lower := strings.ToLower(n.Name)
				switch lower {
				case "reset":
					local = c.Style{}
				case "br", "newline":
					runes = append(runes, '\n')
					styles = append(styles, local)
				case "gradient":
					// Nested gradient: resolve it fully, then re-flatten its
					// own output so it still participates correctly here.
					for _, comp := range renderGradient(n, local) {
						if txt, ok := comp.(*c.Text); ok {
							for _, r := range txt.Content {
								runes = append(runes, r)
								styles = append(styles, txt.S)
							}
						}
					}
				default:
					walk(n.Children, applyTag(local, n))
				}
			}
		}
	}
	walk(nodes, inherited)
	return runes, styles
}

func sameStyle(a, b c.Style) bool {
	return colorEq(a.Color, b.Color) &&
		a.Bold == b.Bold && a.Italic == b.Italic &&
		a.Underlined == b.Underlined && a.Strikethrough == b.Strikethrough &&
		a.Obfuscated == b.Obfuscated &&
		a.ClickEvent == b.ClickEvent && a.HoverEvent == b.HoverEvent
}

func colorEq(a, b mccolor.Color) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Hex() == b.Hex()
}

// --- Gradient math --------------------------------------------------------

type rgb struct{ r, g, b byte }

func hexRGB(v uint32) rgb {
	return rgb{byte(v >> 16), byte(v >> 8), byte(v)}
}

func rgbToColor(c rgb) mccolor.Color {
	return mccolor.HexInt(int(c.r)<<16 | int(c.g)<<8 | int(c.b))
}

func lerpRGB(a, b rgb, t float64) rgb {
	return rgb{lerpByte(a.r, b.r, t), lerpByte(a.g, b.g, t), lerpByte(a.b, b.b, t)}
}

func lerpByte(a, b byte, t float64) byte {
	return byte(math.Round(float64(a) + (float64(b)-float64(a))*t))
}

// gradientCalc mirrors GradientTag's init()/color() logic exactly.
type gradientCalc struct {
	colors     []rgb
	multiplier float64
	phase      float64 // already scaled into color-index space
}

func newGradientCalc(colors []rgb, rawPhase float64, size int) gradientCalc {
	cs := make([]rgb, len(colors))
	copy(cs, colors)

	phase := rawPhase
	if rawPhase < 0 {
		// [-1, 0) -> [0, 1), and the gradient runs in reverse.
		for i, j := 0, len(cs)-1; i < j; i, j = i+1, j-1 {
			cs[i], cs[j] = cs[j], cs[i]
		}
		phase = 1 + rawPhase
	}

	multiplier := 0.0
	if size > 1 {
		multiplier = float64(len(cs)-1) / float64(size-1)
	}
	phase *= float64(len(cs) - 1)

	return gradientCalc{colors: cs, multiplier: multiplier, phase: phase}
}

func (g gradientCalc) at(index int) rgb {
	n := len(g.colors)
	position := float64(index)*g.multiplier + g.phase
	lowF := math.Floor(position)
	high := int(math.Ceil(position)) % n
	low := int(lowF) % n
	frac := position - lowF
	return lerpRGB(g.colors[low], g.colors[high], frac)
}

// parseGradientArgs splits gradient args into colors and an optional trailing
// numeric phase (defaulting to 0).
func parseGradientArgs(args []string) (colors []rgb, phase float64) {
	for i, a := range args {
		if i == len(args)-1 {
			if f, err := strconv.ParseFloat(a, 64); err == nil {
				phase = f
				continue
			}
		}
		if col, ok := resolveColor(a); ok {
			colors = append(colors, hexToRGB(col.Hex()))
		}
	}
	return colors, phase
}

func hexToRGB(hex string) rgb {
	hex = strings.TrimPrefix(hex, "#")
	v, _ := strconv.ParseUint(hex, 16, 32)
	return hexRGB(uint32(v))
}
