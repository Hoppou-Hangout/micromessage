package micromessage

import (
	"math"
	"strconv"
	"strings"

	mccolor "go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/key"
	"go.minekube.com/common/minecraft/nbt"

	"github.com/google/uuid"
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

	if lower == "shadow" {
		if negate {
			// <!shadow> explicitly clears any inherited shadow color.
			style.ShadowColor = &c.ShadowColor{}
			return style
		}
		if sc, ok := parseShadowColor(node.Args); ok {
			style.ShadowColor = sc
		}
		return style
	}

	switch lower {
	case "insert":
		if len(node.Args) > 0 {
			v := strings.Join(node.Args, ":")
			style.Insertion = &v
		}
		return style

	case "font":
		if k := parseKeyArgs(node.Args); k != nil {
			style.Font = k
		}
		return style
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
		if ev := buildHoverEvent(node.Args); ev != nil {
			style.HoverEvent = ev
		}
		return style
	}

	// Bare color tags, e.g. <red> or <#ff0000>.
	if col, ok := resolveColor(name); ok {
		style.Color = col
	}
	return style
}

// parseKeyString turns "value" (default "minecraft" namespace) or a single
// "namespace:value" argument (only reachable via a quoted arg, since bare
// colons are otherwise split into separate tag arguments) into a key.Key.
func parseKeyString(s string) key.Key {
	if ns, v, ok := strings.Cut(s, ":"); ok {
		return key.New(ns, v)
	}
	return key.New(key.MinecraftNamespace, s)
}

// parseKeyArgs handles <font:value> and <font:namespace:value>, the two
// unquoted forms MiniMessage accepts for a key argument split across
// separate tag args.
func parseKeyArgs(args []string) key.Key {
	switch len(args) {
	case 1:
		return parseKeyString(args[0])
	case 2:
		return key.New(args[0], args[1])
	default:
		return nil
	}
}

// parseShadowColor implements <shadow:color[:alpha]>, accepting either an
// explicit #RRGGBBAA literal or a named/hex color plus an optional alpha
// float in [0,1] (defaulting to 0.25, matching Adventure's default).
func parseShadowColor(args []string) (*c.ShadowColor, bool) {
	if len(args) == 0 {
		return nil, false
	}
	arg := args[0]
	if strings.HasPrefix(arg, "#") && len(arg) == 9 {
		v, err := strconv.ParseUint(arg[1:], 16, 32)
		if err != nil {
			return nil, false
		}
		r, g, b, a := byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
		return &c.ShadowColor{ARGB: uint32(a)<<24 | uint32(r)<<16 | uint32(g)<<8 | uint32(b)}, true
	}

	col, ok := resolveColor(arg)
	if !ok {
		return nil, false
	}
	alpha := 0.25
	if len(args) > 1 {
		if f, err := strconv.ParseFloat(args[1], 64); err == nil {
			alpha = f
		}
	}
	rv := hexToRGB(col.Hex())
	a := byte(math.Round(alpha * 255))
	return &c.ShadowColor{ARGB: uint32(a)<<24 | uint32(rv.r)<<16 | uint32(rv.g)<<8 | uint32(rv.b)}, true
}

// buildHoverEvent implements <hover:show_text:VALUE>, <hover:show_item:ID[:COUNT[:NBT]]>
// and <hover:show_entity:TYPE:UUID[:NAME]>. Namespaced IDs (item/entity type)
// need quoting (e.g. "minecraft:diamond") since a bare colon is otherwise
// consumed as a tag-argument separator; unquoted, the ID is assumed to be in
// the "minecraft" namespace.
func buildHoverEvent(args []string) c.HoverEvent {
	if len(args) == 0 {
		return nil
	}
	action, rest := strings.ToLower(args[0]), args[1:]

	switch action {
	case "show_text":
		return c.ShowText(renderInline(strings.Join(rest, ":")))

	case "show_item":
		if len(rest) == 0 {
			return nil
		}
		item := &c.ShowItemHoverType{Item: parseKeyString(rest[0]), Count: 1}
		if len(rest) > 1 {
			if n, err := strconv.Atoi(rest[1]); err == nil {
				item.Count = n
			}
		}
		if len(rest) > 2 {
			item.NBT = nbt.NewBinaryTagHolder(strings.Join(rest[2:], ":"))
		}
		return c.ShowItem(item)

	case "show_entity":
		if len(rest) < 2 {
			return nil
		}
		id, err := uuid.Parse(rest[1])
		if err != nil {
			return nil
		}
		entity := &c.ShowEntityHoverType{Type: parseKeyString(rest[0]), Id: id}
		if len(rest) > 2 {
			entity.Name = renderInline(strings.Join(rest[2:], ":"))
		}
		return c.ShowEntity(entity)

	default:
		return nil
	}
}

// renderInline parses src as its own MiniMessage document, for values (hover
// text, translation placeholders) that Adventure recursively deserializes.
// Falls back to literal text if src doesn't parse.
func renderInline(src string) *c.Text {
	p, err := newParser(src)
	if err != nil {
		return &c.Text{Content: src}
	}
	nodes, err := p.Parse()
	if err != nil {
		return &c.Text{Content: src}
	}
	comps := render(nodes, c.Style{})
	// Unwrap the common case of plain, unstyled text (no tags) into a single
	// flat *c.Text, matching what real MiniMessage produces for such input
	// instead of an empty-content root with one child.
	if len(comps) == 1 {
		if txt, ok := comps[0].(*c.Text); ok && len(txt.Extra) == 0 {
			return txt
		}
	}
	return &c.Text{Extra: comps}
}

// buildTranslation implements <lang:key:with...>/<tr>/<translate> and their
// <..._or:key:fallback:with...> fallback-carrying variants.
func buildTranslation(node *Node, cur c.Style) *c.Translation {
	if len(node.Args) == 0 {
		return nil
	}
	tKey, rest := node.Args[0], node.Args[1:]

	var fallback string
	if strings.HasSuffix(strings.ToLower(node.Name), "_or") {
		if len(rest) == 0 {
			return nil
		}
		fallback, rest = rest[0], rest[1:]
	}

	var with []c.Component
	for _, a := range rest {
		with = append(with, renderInline(a))
	}
	return &c.Translation{Key: tKey, S: cur, With: with, Fallback: fallback}
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
			case "lang", "tr", "translate", "lang_or", "tr_or", "translate_or":
				if tr := buildTranslation(n, cur); tr != nil {
					out = append(out, tr)
				}
			case "gradient":
				out = append(out, renderGradient(n, cur)...)
			case "rainbow":
				out = append(out, renderRainbow(n, cur)...)
			case "transition":
				out = append(out, renderTransition(n, cur)...)
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

// renderRainbow implements <rainbow:[!][phase]>, a faithful port of
// Adventure's RainbowTag: hue cycles once across the flattened character
// span, "!" reverses direction, and phase (tenths, e.g. "5" -> 0.5) shifts
// the starting hue.
func renderRainbow(node *Node, inherited c.Style) []c.Component {
	reversed, phase := parseRainbowArgs(node.Args)

	base := inherited
	base.Color = nil
	runes, styles := flattenChars(node.Children, base)
	n := len(runes)

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
		idx := i
		if reversed {
			idx = n - 1 - i
		}
		hue := math.Mod(float64(idx)/float64(n)+phase, 1)
		if hue < 0 {
			hue += 1
		}

		s := styles[i]
		if s.Color == nil {
			s.Color = rgbToColor(hsvToRGB(hue, 1, 1))
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

// parseRainbowArgs parses the single optional <rainbow> argument: an
// optional leading "!" (reverse) followed by an optional integer phase in
// tenths (matching Adventure, where "<rainbow:5>" means phase 0.5).
func parseRainbowArgs(args []string) (reversed bool, phase float64) {
	if len(args) == 0 {
		return false, 0
	}
	v := args[0]
	if strings.HasPrefix(v, "!") {
		reversed = true
		v = v[1:]
	}
	if v == "" {
		return reversed, 0
	}
	if n, err := strconv.Atoi(v); err == nil {
		phase = float64(n) / 10
	}
	return reversed, phase
}

// hsvToRGB converts an HSV color (h, s, v all in [0,1]) to RGB bytes.
func hsvToRGB(h, s, v float64) rgb {
	i := int(h * 6)
	f := h*6 - float64(i)
	p := v * (1 - s)
	q := v * (1 - f*s)
	t := v * (1 - (1-f)*s)

	var r, g, b float64
	switch i % 6 {
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
	return rgb{
		byte(math.Round(r * 255)),
		byte(math.Round(g * 255)),
		byte(math.Round(b * 255)),
	}
}

// renderTransition implements <transition:c1:c2:...:cN[:phase]>. Unlike
// <gradient>, real MiniMessage's TransitionTag computes a *single* color
// from the phase (it's meant to be animated by re-resolving the tag over
// time via a placeholder), not a per-character interpolation - so here it's
// equivalent to a plain color wrap using that computed color. A faithful
// port of TransitionTag.color().
func renderTransition(node *Node, inherited c.Style) []c.Component {
	colors, phase := parseGradientArgs(node.Args)
	if len(colors) < 2 {
		colors = []rgb{hexRGB(0xffffff), hexRGB(0x000000)}
	}

	style := inherited
	style.Color = rgbToColor(transitionColor(colors, phase))
	return render(node.Children, style)
}

// transitionColor is a direct port of TransitionTag.color(): colors is
// mutated in place (reversed) when phase is negative, matching the Java
// implementation's constructor-time reversal.
func transitionColor(colors []rgb, phase float64) rgb {
	negative := phase < 0
	if negative {
		phase = 1 + phase
		for i, j := 0, len(colors)-1; i < j; i, j = i+1, j-1 {
			colors[i], colors[j] = colors[j], colors[i]
		}
	}

	steps := 1 / float64(len(colors)-1)
	for i := 1; i < len(colors); i++ {
		val := float64(i) * steps
		if val >= phase {
			factor := 1 + (phase-val)*float64(len(colors)-1)
			if negative {
				return lerpRGB(colors[i], colors[i-1], 1-factor)
			}
			return lerpRGB(colors[i-1], colors[i], factor)
		}
	}
	return colors[0]
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
				case "gradient", "rainbow", "transition":
					// Nested gradient/rainbow/transition: resolve it fully,
					// then re-flatten its own output so it still
					// participates correctly here.
					var comps []c.Component
					switch lower {
					case "gradient":
						comps = renderGradient(n, local)
					case "rainbow":
						comps = renderRainbow(n, local)
					case "transition":
						comps = renderTransition(n, local)
					}
					for _, comp := range comps {
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
		a.ClickEvent == b.ClickEvent && a.HoverEvent == b.HoverEvent &&
		shadowEq(a.ShadowColor, b.ShadowColor) &&
		insertionEq(a.Insertion, b.Insertion) &&
		fontEq(a.Font, b.Font)
}

func shadowEq(a, b *c.ShadowColor) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.ARGB == b.ARGB
}

func insertionEq(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func fontEq(a, b key.Key) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.String() == b.String()
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
