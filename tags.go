package micromessage

import (
	"strconv"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/key"
)

// Maps every accepted spelling to its corresponding decoration name.
var decorationAliases = map[string]string{
	"b":             Bold,
	"bold":          Bold,
	"i":             Italic,
	"em":            Italic,
	"italic":        Italic,
	"u":             Underlined,
	"underlined":    Underlined,
	"st":            Strikethrough,
	"strikethrough": Strikethrough,
	"obf":           Obfuscated,
	"obfuscated":    Obfuscated,
}

func canonicalDecoration(name string) (string, bool) {
	n, ok := decorationAliases[strings.ToLower(name)]
	return n, ok
}

// resolveTagStyle resolves a tag into a minekube component.Style.
// It returns (style, ok, err): ok is false if the tag is not a recognized
// style tag.
func resolveTagStyle(name string, args []string, deserialize func(string) (*component.Text, error)) (component.Style, bool, error) {
	style := component.Style{}

	lower := strings.ToLower(name)

	// Negated decoration shorthand: <!bold>
	if strings.HasPrefix(name, "!") {
		if dec, ok := canonicalDecoration(name[1:]); ok {
			style.SetDecoration(component.Decoration(dec), component.False)
			return style, true, nil
		}
		return style, false, nil
	}

	// Decoration tags (with optional explicit true/false argument)
	if dec, ok := canonicalDecoration(lower); ok {
		val := component.True
		if len(args) > 0 {
			switch strings.ToLower(args[0]) {
			case "false":
				val = component.False
			case "true":
				val = component.True
			}
		}
		style.SetDecoration(component.Decoration(dec), val)
		return style, true, nil
	}

	if lower == "color" || lower == "colour" || lower == "c" {
		if len(args) == 0 {
			return style, false, nil
		}
		col := ResolveColor(args[0])
		if col == nil {
			return style, false, nil
		}
		style.Color = (*mkRGB)(col)
		return style, true, nil
	}

	if col := ResolveColor(name); col != nil {
		style.Color = (*mkRGB)(col)
		return style, true, nil
	}

	switch lower {
	case "click":
		if len(args) < 2 {
			return style, false, nil
		}
		action := component.ClickActions[strings.ToLower(args[0])]
		if action == nil {
			return style, false, nil
		}
		style.ClickEvent = component.NewClickEvent(action, args[1])
		return style, true, nil

	case "hover":
		if len(args) < 2 {
			return style, false, nil
		}
		action := strings.ToLower(args[0])
		if action != "show_text" {
			return style, true, nil
		}
		comp, err := deserialize(args[1])
		if err != nil {
			return style, false, err
		}
		style.HoverEvent = component.ShowText(comp)
		return style, true, nil

	case "insert", "insertion":
		if len(args) < 1 {
			return style, false, nil
		}
		v := args[0]
		style.Insertion = &v
		return style, true, nil

	case "font":
		if len(args) < 1 {
			return style, false, nil
		}
		fontStr := args[0]
		if len(args) >= 2 {
			fontStr = args[0] + ":" + args[1]
		}
		if k, err := key.Parse(fontStr); err == nil {
			style.Font = k
		}
		return style, true, nil
	}

	return style, false, nil
}

func isNewlineTag(name string) bool {
	l := strings.ToLower(name)
	return l == "newline" || l == "br"
}

func isKnownTagName(name string) bool {
	lower := strings.ToLower(name)
	if strings.HasPrefix(name, "!") {
		_, ok := canonicalDecoration(name[1:])
		return ok
	}
	if _, ok := canonicalDecoration(lower); ok {
		return true
	}
	if lower == "color" || lower == "colour" || lower == "c" {
		return true
	}
	if ResolveColor(name) != nil {
		return true
	}
	switch lower {
	case "click", "hover", "insert", "insertion", "font", "reset", "newline", "br", "gradient", "rainbow":
		return true
	}
	return false
}

// Outputs a sequence of colors for gradient/rainbow tags.
type colorAdvancer interface {
	init(size int)
	color() *colorful.Color
	advance()
}

// This struct implements <gradient:...> color progression.
type gradientAdvancer struct {
	colors     []*colorful.Color
	phase      float64
	multiplier float64
	index      int
}

func newGradientAdvancer(colors []*colorful.Color, phase float64) *gradientAdvancer {
	if len(colors) == 0 {
		colors = []*colorful.Color{RGB(0xff, 0xff, 0xff), RGB(0, 0, 0)}
	}
	g := &gradientAdvancer{colors: colors}
	if phase < 0 {
		reversed := make([]*colorful.Color, len(colors))
		for i, c := range colors {
			reversed[len(colors)-1-i] = c
		}
		g.colors = reversed
		g.phase = 1 + phase
	} else {
		g.phase = phase
	}
	return g
}

func (g *gradientAdvancer) init(size int) {
	if size == 1 {
		g.multiplier = 0
	} else {
		g.multiplier = float64(len(g.colors)-1) / float64(size-1)
	}
	g.phase *= float64(len(g.colors) - 1)
	g.index = 0
}

func (g *gradientAdvancer) advance() { g.index++ }

func (g *gradientAdvancer) color() *colorful.Color {
	position := float64(g.index)*g.multiplier + g.phase
	low := int(floor(position))
	high := int(ceil(position)) % len(g.colors)
	lowMod := ((low % len(g.colors)) + len(g.colors)) % len(g.colors)
	t := position - floor(position)
	return LerpColor(t, g.colors[lowMod], g.colors[high])
}

func floor(f float64) float64 {
	i := int(f)
	if f < 0 && float64(i) != f {
		i--
	}
	return float64(i)
}

func ceil(f float64) float64 {
	i := int(f)
	if f > 0 && float64(i) != f {
		i++
	}
	return float64(i)
}

// This implements <rainbow:...> color progression.
type rainbowAdvancer struct {
	reversed     bool
	dividedPhase float64
	size         int
	colorIndex   int
}

func newRainbowAdvancer(reversed bool, phase int) *rainbowAdvancer {
	return &rainbowAdvancer{reversed: reversed, dividedPhase: float64(phase) / 10}
}

func (r *rainbowAdvancer) init(size int) {
	r.size = size
	if r.reversed {
		r.colorIndex = size - 1
	}
}

func (r *rainbowAdvancer) advance() {
	if r.reversed {
		if r.colorIndex == 0 {
			r.colorIndex = r.size - 1
		} else {
			r.colorIndex--
		}
	} else {
		r.colorIndex++
	}
}

func (r *rainbowAdvancer) color() *colorful.Color {
	hue := (float64(r.colorIndex)/float64(r.size) + r.dividedPhase)
	hue -= floor(hue)
	return HSVColor(hue, 1, 1)
}

func parseGradientArgs(args []string) (colors []*colorful.Color, phase float64, err error) {
	if len(args) == 0 {
		return nil, 0, nil
	}
	for i, a := range args {
		if c := ResolveColor(a); c != nil {
			colors = append(colors, c)
			continue
		}
		if i == len(args)-1 {
			if p, ok := argAsFloat64(a); ok {
				phase = p
				continue
			}
		}
		return nil, 0, &ParseError{Message: "Unable to parse a color from '" + a + "'"}
	}
	return colors, phase, nil
}

func parseRainbowArgs(args []string) (reversed bool, phase int, err error) {
	if len(args) == 0 {
		return false, 0, nil
	}
	v := args[0]
	if strings.HasPrefix(v, "!") {
		reversed = true
		v = v[1:]
	}
	if v == "" {
		return reversed, 0, nil
	}
	p, e := strconv.Atoi(v)
	if e != nil {
		return false, 0, &ParseError{Message: "Expected phase, got " + v}
	}
	return reversed, p, nil
}

func isGradientTag(name string) bool { return strings.EqualFold(name, "gradient") }
func isRainbowTag(name string) bool  { return strings.EqualFold(name, "rainbow") }

func totalTextLen(comp *component.Text) int {
	n := codePointLen(comp.Content)
	for _, c := range comp.Extra {
		n += totalTextLen(asText(c))
	}
	return n
}

func applyColorChanging(comp *component.Text, adv colorAdvancer) *component.Text {
	adv.init(totalTextLen(comp))
	return applyColorRec(comp, adv)
}

func applyColorRec(comp *component.Text, adv colorAdvancer) *component.Text {
	if comp.S.Color != nil {
		n := totalTextLen(comp)
		for i := 0; i < n; i++ {
			adv.advance()
		}
		return cloneText(comp)
	}

	if comp.Content != "" {
		runes := []rune(comp.Content)
		out := &component.Text{S: cloneStyle(comp.S)}
		out.S.Color = nil
		for _, r := range runes {
			col := adv.color()
			adv.advance()
			chStyle := cloneStyle(comp.S)
			chStyle.Color = (*mkRGB)(col)
			out.Extra = append(out.Extra, &component.Text{
				Content: string(r),
				S:       chStyle,
			})
		}
		for _, ch := range comp.Extra {
			out.Extra = append(out.Extra, applyColorRec(asText(ch), adv))
		}
		return out
	}

	out := &component.Text{S: cloneStyle(comp.S)}
	for _, ch := range comp.Extra {
		out.Extra = append(out.Extra, applyColorRec(asText(ch), adv))
	}
	return out
}