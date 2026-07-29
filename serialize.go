package micromessage

import "strings"

// PlainText flattens a Component tree to its raw text content, dropping all styling.
func PlainText(c *Component) string {
	var sb strings.Builder
	writePlainText(c, &sb)
	return sb.String()
}

func writePlainText(c *Component, sb *strings.Builder) {
	if c == nil {
		return
	}
	sb.WriteString(c.Text)
	for _, ch := range c.Children {
		writePlainText(ch, sb)
	}
}

// This renders a component tree back into minimessage source
func Serialize(c *Component) string {
	var sb strings.Builder
	serializeNode(c, NewStyle(), &sb)
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

func serializeNode(c *Component, parent Style, sb *strings.Builder) {
	if c == nil {
		return
	}

	tags := diffStyleTags(parent, c.Style)
	for _, t := range tags {
		sb.WriteByte('<')
		sb.WriteString(t)
		sb.WriteByte('>')
	}

	sb.WriteString(escapeLiteral(c.Text))

	for _, ch := range c.Children {
		serializeNode(ch, c.Style, sb)
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
func diffStyleTags(parent, child Style) []string {
	var tags []string

	if child.Color != nil && !child.Color.Equal(parent.Color) {
		if name := NamedColorName(child.Color); name != "" {
			tags = append(tags, name)
		} else {
			tags = append(tags, child.Color.HexString())
		}
	}

	for _, dec := range AllDecorations {
		cv := child.Decorations[dec]
		pv := parent.Decorations[dec]
		if cv != Unset && cv != pv {
			if cv == True {
				tags = append(tags, dec)
			} else {
				tags = append(tags, "!"+dec)
			}
		}
	}

	if child.ClickEvent != nil && (parent.ClickEvent == nil || *child.ClickEvent != *parent.ClickEvent) {
		tags = append(tags, "click:"+string(child.ClickEvent.Action)+":"+child.ClickEvent.Value)
	}

	if child.HoverEvent != nil && child.HoverEvent != parent.HoverEvent {
		if child.HoverEvent.Action == ShowText && child.HoverEvent.Value != nil {
			tags = append(tags, "hover:show_text:'"+Serialize(child.HoverEvent.Value)+"'")
		}
	}

	if child.Insertion != nil && (parent.Insertion == nil || *child.Insertion != *parent.Insertion) {
		tags = append(tags, "insert:"+*child.Insertion)
	}

	if child.Font != nil && (parent.Font == nil || *child.Font != *parent.Font) {
		tags = append(tags, "font:"+*child.Font)
	}

	return tags
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
