// Package micromessage is a MiniMessage (https://docs.advntr.dev/minimessage)
// parser and renderer for Gate/Minekube's component model
// (go.minekube.com/common/minecraft/component).
package micromessage

import (
	c "go.minekube.com/common/minecraft/component"
)

// Deserialize parses a MiniMessage string into a *component.Text tree
// compatible with go.minekube.com/common and, by extension, Gate.
//
// By default every standard tag (StandardTags.All()) is recognized and
// unresolved/invalid tags are rendered as literal text -- matching
// Adventure's MiniMessage.miniMessage(). Pass Options to customize this:
// WithTags to restrict/replace the tag vocabulary, WithTagResolver to add
// placeholders or other dynamic tags, WithPreprocessor to rewrite the raw
// input before parsing, WithStrict to error on unclosed tags, and WithDebug
// to receive diagnostics.
func Deserialize(input string, opts ...Option) (result *c.Text, err error) {
	defer func() {
		if r := recover(); r != nil {
			if ae, ok := r.(*ArgumentError); ok {
				err = ae
				return
			}
			panic(r)
		}
	}()

	var o options
	for _, opt := range opts {
		opt(&o)
	}
	for _, pp := range o.preprocessors {
		input = pp(input)
	}

	resolvers := o.resolvers()

	nodes, err := parse(input, resolvers, o.strict)
	if err != nil {
		return nil, err
	}

	result = &c.Text{Extra: render(nodes, c.Style{}, resolvers, 0)}
	for _, pp := range o.postProcess {
		result = pp(result)
	}
	return result, nil
}

// MustDeserialize is similar to Deserialize but panics on error.
func MustDeserialize(input string, opts ...Option) *c.Text {
	t, err := Deserialize(input, opts...)
	if err != nil {
		panic(err)
	}
	return t
}
