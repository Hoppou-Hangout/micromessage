package micromessage

import (
	"fmt"
	"strings"
)

type Placeholder struct {
	Name      string
	Component *Component
}

func TextPlaceholder(name, miniMessage string) Placeholder {
	c, _ := Deserialize(miniMessage)
	return Placeholder{Name: name, Component: c}
}

func ComponentPlaceholder(name string, c *Component) Placeholder {
	return Placeholder{Name: name, Component: c}
}

func StringPlaceholder(name, value string) Placeholder {
	return Placeholder{Name: name, Component: Text(value)}
}

func findPlaceholder(placeholders []Placeholder, name string) (*Component, bool) {
	for _, p := range placeholders {
		if strings.EqualFold(p.Name, name) {
			return p.Component, true
		}
	}
	return nil, false
}

func Deserialize(input string, placeholders ...Placeholder) (*Component, error) {
	return deserialize(input, false, placeholders)
}

func DeserializeStrict(input string, placeholders ...Placeholder) (*Component, error) {
	return deserialize(input, true, placeholders)
}

func deserialize(input string, strict bool, placeholders []Placeholder) (*Component, error) {
	root, err := parseTree(input, strict)
	if err != nil {
		return nil, err
	}

	deserializeSub := func(s string) (*Component, error) {
		return deserialize(s, strict, placeholders)
	}

	comp, err := buildComponent(root, placeholders, deserializeSub, strict)
	if err != nil {
		return nil, err
	}
	if comp == nil {
		comp = Empty()
	}
	return comp, nil
}

func buildComponent(n *elementNode, placeholders []Placeholder, deserializeSub func(string) (*Component, error), strict bool) (*Component, error) {
	if n.isText {
		return Text(n.text), nil
	}

	if n.tagName == "" {
		// root or plain container: fold children together
		return foldChildren(n, placeholders, deserializeSub, strict)
	}

	// Placeholder substitution
	if repl, ok := findPlaceholder(placeholders, n.tagName); ok {
		out := repl.Clone()
		if len(n.children) > 0 {
			kids, err := buildChildren(n, placeholders, deserializeSub, strict)
			if err != nil {
				return nil, err
			}
			out.Children = append(out.Children, kids...)
		}
		return out, nil
	}

	if isNewlineTag(n.tagName) {
		out := Text("\n")
		if len(n.children) > 0 {
			kids, err := buildChildren(n, placeholders, deserializeSub, strict)
			if err != nil {
				return nil, err
			}
			out.Children = append(out.Children, kids...)
		}
		return out, nil
	}

	if isGradientTag(n.tagName) || isRainbowTag(n.tagName) {
		inner, err := foldChildren(n, placeholders, deserializeSub, strict)
		if err != nil {
			return nil, err
		}
		var adv colorAdvancer
		if isGradientTag(n.tagName) {
			colors, phase, perr := parseGradientArgs(n.tagArgs)
			if perr != nil {
				if strict {
					return nil, perr
				}
				return literalTag(n), nil
			}
			adv = newGradientAdvancer(colors, phase)
		} else {
			reversed, phase, perr := parseRainbowArgs(n.tagArgs)
			if perr != nil {
				if strict {
					return nil, perr
				}
				return literalTag(n), nil
			}
			adv = newRainbowAdvancer(reversed, phase)
		}
		return applyColorChanging(inner, adv), nil
	}

	style, ok, err := resolveTagStyle(n.tagName, n.tagArgs, deserializeSub)
	if err != nil {
		if strict {
			return nil, err
		}
		return literalTag(n), nil
	}
	if !ok {
		// Render component literally as text if tag is unknown
		if strict {
			return nil, &ParseError{Message: fmt.Sprintf("Unknown tag '%s'", n.tagName)}
		}
		return literalTagWithChildren(n, placeholders, deserializeSub, strict)
	}

	out := &Component{Style: style}
	kids, err := buildChildren(n, placeholders, deserializeSub, strict)
	if err != nil {
		return nil, err
	}
	out.Children = kids
	return out, nil
}

func buildChildren(n *elementNode, placeholders []Placeholder, deserializeSub func(string) (*Component, error), strict bool) ([]*Component, error) {
	var out []*Component
	for _, c := range n.children {
		cc, err := buildComponent(c, placeholders, deserializeSub, strict)
		if err != nil {
			return nil, err
		}
		out = append(out, cc)
	}
	return out, nil
}

// Builds all children into a single container Component.
func foldChildren(n *elementNode, placeholders []Placeholder, deserializeSub func(string) (*Component, error), strict bool) (*Component, error) {
	kids, err := buildChildren(n, placeholders, deserializeSub, strict)
	if err != nil {
		return nil, err
	}
	out := Empty()
	out.Children = kids
	return out, nil
}

// Renders an unresolved self closing tag as literal text "<name:args>".
func literalTag(n *elementNode) *Component {
	return Text(renderTagLiteral(n))
}

// Renders an unresolved tag as literal opening text
func literalTagWithChildren(n *elementNode, placeholders []Placeholder, deserializeSub func(string) (*Component, error), strict bool) (*Component, error) {
	out := Empty()
	out.Children = append(out.Children, Text(renderTagLiteral(n)))
	kids, err := buildChildren(n, placeholders, deserializeSub, strict)
	if err != nil {
		return nil, err
	}
	out.Children = append(out.Children, kids...)
	return out, nil
}

func renderTagLiteral(n *elementNode) string {
	var sb strings.Builder
	sb.WriteByte('<')
	sb.WriteString(n.tagName)
	for _, a := range n.tagArgs {
		sb.WriteByte(':')
		sb.WriteString(a)
	}
	sb.WriteByte('>')
	return sb.String()
}
