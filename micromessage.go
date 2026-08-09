// Package micromessage is a MiniMessage (https://docs.advntr.dev/minimessage)
// parser and renderer for Gate/Minekube's component model
// (go.minekube.com/common/minecraft/component).
package micromessage

import (
	c "go.minekube.com/common/minecraft/component"
)

// Deserialize parses a MiniMessage string into a *component.Text tree
// compatible with go.minekube.com/common and, by extension, Gate.
func Deserialize(input string) (*c.Text, error) {
	p, err := newParser(input)
	if err != nil {
		return nil, err
	}
	nodes, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return &c.Text{Extra: render(nodes, c.Style{})}, nil
}

// MustDeserialize is similar to Deserialize but panics on error.
func MustDeserialize(input string) *c.Text {
	t, err := Deserialize(input)
	if err != nil {
		panic(err)
	}
	return t
}
