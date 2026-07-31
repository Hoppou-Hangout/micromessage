package micromessage

import (
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"go.minekube.com/common/minecraft/component"
	mcolor "go.minekube.com/common/minecraft/color"
)

// PlainText flattens a Component tree to its raw text content, dropping all styling.
func PlainText(c *Component) string {
	if c == nil {
		return ""
	}
	return plainTextMK(componentToMinekube(c))
}

func plainTextMK(c *component.Text) string {
	if c == nil {
		return ""
	}
	var sb strings.Builder
	writePlainTextMK(c, &sb)
	return sb.String()
}

func writePlainTextMK(c *component.Text, sb *strings.Builder) {
	if c == nil {
		return
	}
	sb.WriteString(c.Content)
	for _, ch := range c.Extra {
		writePlainTextMK(asText(ch), sb)
	}
}

// This renders a component tree back into minimessage source
func Serialize(c *Component) string {
	if c == nil {
		return ""
	}
	var sb strings.Builder
	serializeNodeMK(componentToMinekube(c), component.Style{}, &sb)
	return trimTrailingCloseTags(sb.String())
}

func trimTrailingCloseTags(s string) string {
	for strings.HasSuffix(s, ">") {
		idx := strings.LastIndex(s, "</")
		if idx == -1 || idx != strings.LastIndex(s, "<") {
			break
		}
		s = s[:idx]
	}
	return s
}

func serializeNodeMK(c *component.Text, parent component.Style, sb *strings.Builder) {
	if c == nil {
		return
	}

	tags := diffStyleTagsMK(parent, c.S)
	for _, t := range tags {
		sb.WriteByte('<')
		sb.WriteString(t)
		sb.WriteByte('>')
	}

	sb.WriteString(escapeLiteral(c.Content))

	for _, ch := range c.Extra {
		serializeNodeMK(asText(ch), c.S, sb)
	}

	for i := len(tags) - 1; i >= 0; i-- {
		name := tags[i]
		if idx := strings.IndexByte(name, ':'); idx >= 0 {
			name = name[:idx]
		}
		sb.WriteString("</")
		sb.WriteString(name)
		sb.WriteByte('>')
	}
}

// Returns opening tags needed to go from parent style to child style.
func diffStyleTagsMK(parent, child component.Style) []string {
	var tags []string

	if child.Color != nil && !colorsEqualMK(parent.Color, child.Color) {
		col := colorfulFromMK(child.Color)
		if name := NamedColorName(col); name != "" {
			tags = append(tags, name)
		} else {
			tags = append(tags, col.Hex())
		}
	}

	for _, dec := range AllDecorations {
		d := component.Decoration(dec)
		cv := child.Decoration(d)
		pv := parent.Decoration(d)
		if cv != component.NotSet && cv != pv {
			if cv == component.True {
				tags = append(tags, dec)
			} else {
				tags = append(tags, "!"+dec)
			}
		}
	}

	if child.ClickEvent != nil && (parent.ClickEvent == nil || !clickEventsEqualMK(parent.ClickEvent, child.ClickEvent)) {
		tags = append(tags, "click:"+child.ClickEvent.Action().Name()+":"+child.ClickEvent.Value())
	}

	if child.HoverEvent != nil && child.HoverEvent != parent.HoverEvent {
		if child.HoverEvent.Action().Name() == "show_text" {
			if val, ok := child.HoverEvent.Value().(component.Component); ok {
				tags = append(tags, "hover:show_text:'"+plainTextMK(asText(val))+"'")
			}
		}
	}

	if child.Insertion != nil && (parent.Insertion == nil || *child.Insertion != *parent.Insertion) {
		tags = append(tags, "insert:"+*child.Insertion)
	}

	if child.Font != nil && (parent.Font == nil || child.Font.String() != parent.Font.String()) {
		tags = append(tags, "font:"+child.Font.String())
	}

	return tags
}

func colorsEqualMK(a, b mcolor.Color) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	return ar == br && ag == bg && ab == bb
}

func clickEventsEqualMK(a, b component.ClickEvent) bool {
	return a.Action().Name() == b.Action().Name() && a.Value() == b.Value()
}

// colorfulFromMK recovers a *colorful.Color from a minekube color.Color.
func colorfulFromMK(c mcolor.Color) *colorful.Color {
	if c == nil {
		return nil
	}
	r, g, b, _ := c.RGBA()
	col := colorful.Color{
		R: float64(r) / 65535.0,
		G: float64(g) / 65535.0,
		B: float64(b) / 65535.0,
	}
	return &col
}

func escapeLiteral(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r == '<' || r == '\\' {
			sb.WriteByte('\\')
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

func StripTags(input string, placeholders ...Placeholder) string {
	msg := []rune(input)
	tokens := tokenize(msg)

	names := placeholderNameSet(placeholders)

	var sb strings.Builder
	for _, t := range tokens {
		if t.typ == TokenText {
			sb.WriteString(t.text(msg))
			continue
		}
		if !tokenTagRecognized(msg, t, names) {
			sb.WriteString(t.text(msg))
		}
	}
	return sb.String()
}

func EscapeTags(input string, placeholders ...Placeholder) string {
	msg := []rune(input)
	tokens := tokenize(msg)

	names := placeholderNameSet(placeholders)

	var sb strings.Builder
	for _, t := range tokens {
		if t.typ == TokenText {
			sb.WriteString(t.text(msg))
			continue
		}
		if tokenTagRecognized(msg, t, names) {
			sb.WriteByte('\\')
			escapeToken(&sb, msg, t)
		} else {
			sb.WriteString(t.text(msg))
		}
	}
	return sb.String()
}

func escapeToken(sb *strings.Builder, msg []rune, t *token) {
	sb.WriteByte(tagStart)
	if t.typ == TokenCloseTag {
		sb.WriteByte(closeTag)
	}
	for i, c := range t.children {
		if i != 0 {
			sb.WriteByte(separator)
		}
		inner := []rune(c.text(msg))
		innerTokens := tokenize(inner)
		for _, it := range innerTokens {
			if it.typ == TokenText {
				sb.WriteString(it.text(inner))
			} else {
				escapeToken(sb, inner, it)
			}
		}
	}
	sb.WriteByte(tagEnd)
}

func placeholderNameSet(placeholders []Placeholder) map[string]bool {
	names := map[string]bool{}
	for _, p := range placeholders {
		names[strings.ToLower(p.Name)] = true
	}
	return names
}

func tokenTagRecognized(msg []rune, t *token, placeholderNames map[string]bool) bool {
	if len(t.children) == 0 {
		return false
	}
	nameTok := t.children[0]
	name := unquoteAndEscape(msg, nameTok.start, nameTok.end)
	if placeholderNames[strings.ToLower(name)] {
		return true
	}
	return isKnownTagName(name)
}