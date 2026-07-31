package micromessage

import (
	"fmt"
	"strings"

	"go.minekube.com/common/minecraft/component"
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

func findPlaceholder(placeholders []Placeholder, name string) (*component.Text, bool) {
	for _, p := range placeholders {
		if strings.EqualFold(p.Name, name) {
			return componentToMinekube(p.Component), true
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

	deserializeSub := func(s string) (*component.Text, error) {
		c, err := deserialize(s, strict, placeholders)
		if err != nil {
			return nil, err
		}
		return componentToMinekube(c), nil
	}

	comp, err := buildComponent(root, placeholders, deserializeSub, strict)
	if err != nil {
		return nil, err
	}
	if comp == nil {
		comp = &component.Text{}
	}
	return componentFromMinekube(comp), nil
}

func buildComponent(n *elementNode, placeholders []Placeholder, deserializeSub func(string) (*component.Text, error), strict bool) (*component.Text, error) {
	if n.isText {
		return &component.Text{Content: n.text}, nil
	}

	if n.tagName == "" {
		// root or plain container: fold children together
		return foldChildren(n, placeholders, deserializeSub, strict)
	}

	// Placeholder substitution
	if repl, ok := findPlaceholder(placeholders, n.tagName); ok {
		out := cloneText(repl)
		if len(n.children) > 0 {
			kids, err := buildChildren(n, placeholders, deserializeSub, strict)
			if err != nil {
				return nil, err
			}
			out.Extra = append(out.Extra, kids...)
		}
		return out, nil
	}

	if isNewlineTag(n.tagName) {
		out := &component.Text{Content: "\n"}
		if len(n.children) > 0 {
			kids, err := buildChildren(n, placeholders, deserializeSub, strict)
			if err != nil {
				return nil, err
			}
			out.Extra = append(out.Extra, kids...)
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

	out := &component.Text{S: style}
	kids, err := buildChildren(n, placeholders, deserializeSub, strict)
	if err != nil {
		return nil, err
	}
	out.Extra = kids
	return out, nil
}

func buildChildren(n *elementNode, placeholders []Placeholder, deserializeSub func(string) (*component.Text, error), strict bool) ([]component.Component, error) {
	var out []component.Component
	for _, c := range n.children {
		cc, err := buildComponent(c, placeholders, deserializeSub, strict)
		if err != nil {
			return nil, err
		}
		out = append(out, cc)
	}
	return out, nil
}

// Builds all children into a single container component.Text.
func foldChildren(n *elementNode, placeholders []Placeholder, deserializeSub func(string) (*component.Text, error), strict bool) (*component.Text, error) {
	kids, err := buildChildren(n, placeholders, deserializeSub, strict)
	if err != nil {
		return nil, err
	}
	return &component.Text{Extra: kids}, nil
}

// Renders an unresolved self closing tag as literal text "<name:args>".
func literalTag(n *elementNode) *component.Text {
	return &component.Text{Content: renderTagLiteral(n)}
}

// Renders an unresolved tag as literal opening text
func literalTagWithChildren(n *elementNode, placeholders []Placeholder, deserializeSub func(string) (*component.Text, error), strict bool) (*component.Text, error) {
	out := &component.Text{}
	out.Extra = append(out.Extra, &component.Text{Content: renderTagLiteral(n)})
	kids, err := buildChildren(n, placeholders, deserializeSub, strict)
	if err != nil {
		return nil, err
	}
	out.Extra = append(out.Extra, kids...)
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