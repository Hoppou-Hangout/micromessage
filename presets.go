package micromessage

import (
	c "go.minekube.com/common/minecraft/component"
)

// Preset is a pre-built bundle of Options, matching Adventure's
// MiniMessage.Preset. Turn one into an Option with Apply:
//
//	msg, err := micromessage.Deserialize(input, micromessage.NonInteractablePreset.Apply())
type Preset []Option

// Apply turns the preset into a single Option, so it can be combined with
// others: Deserialize(input, SomePreset.Apply(), WithTagResolver(...)).
func (p Preset) Apply() Option {
	return func(o *options) {
		for _, opt := range p {
			opt(o)
		}
	}
}

var (
	// DefaultPreset contains every standard tag with no restrictions --
	// the same as never applying a preset at all, matching Adventure's
	// MiniMessage.Preset.DEFAULT.
	DefaultPreset = Preset{WithTags(StandardTags.All())}

	// NonInteractablePreset disables click events, hover events, and text
	// insertion (the "<click>"/"<hover>"/"<insert>" tags are not part of its
	// tag set at all, so they render as literal text), and additionally
	// strips any interactable style a custom TagResolver might still
	// introduce, matching Adventure's MiniMessage.Preset.NON_INTERACTABLE.
	NonInteractablePreset = Preset{
		WithTags(NewTagResolverBuilder().
			Resolver(StandardTags.Color()).
			Resolver(StandardTags.Decorations()).
			Resolver(StandardTags.Font()).
			Resolver(StandardTags.Shadow()).
			Resolver(StandardTags.Gradient()).
			Resolver(StandardTags.Rainbow()).
			Resolver(StandardTags.Transition()).
			Resolver(StandardTags.Translatable()).
			Resolver(StandardTags.Newline()).
			Resolver(StandardTags.Reset()).
			Build()),
		withPostProcess(stripInteractable),
	}

	// FormattedTextPreset only allows text components and their formatting
	// (color, shadow, font, decoration) -- no click/hover/insertion tags,
	// and its post-processor drops any non-text component (e.g. a
	// Translation a custom TagResolver produced) from the result, matching
	// Adventure's MiniMessage.Preset.FORMATTED_TEXT.
	FormattedTextPreset = Preset{
		WithTags(NewTagResolverBuilder().
			Resolver(StandardTags.Color()).
			Resolver(StandardTags.Decorations()).
			Resolver(StandardTags.Font()).
			Resolver(StandardTags.Shadow()).
			Resolver(StandardTags.Gradient()).
			Resolver(StandardTags.Rainbow()).
			Resolver(StandardTags.Transition()).
			Resolver(StandardTags.Newline()).
			Resolver(StandardTags.Reset()).
			Build()),
		withPostProcess(formattedTextOnly),
	}
)

func withPostProcess(fn func(*c.Text) *c.Text) Option {
	return func(o *options) { o.postProcess = append(o.postProcess, fn) }
}

// stripInteractable clears ClickEvent/HoverEvent/Insertion from root and
// every descendant component, regardless of kind.
func stripInteractable(root *c.Text) *c.Text {
	var walk func(comp c.Component)
	walk = func(comp c.Component) {
		s := comp.Style()
		s.ClickEvent = nil
		s.HoverEvent = nil
		s.Insertion = nil
		for _, ch := range comp.Children() {
			walk(ch)
		}
	}
	walk(root)
	return root
}

// formattedTextOnly strips interactable style like stripInteractable, and
// additionally drops any non-*c.Text component from the tree.
func formattedTextOnly(root *c.Text) *c.Text {
	stripInteractableText(root)
	root.Extra = filterTextChildren(root.Extra)
	return root
}

func stripInteractableText(t *c.Text) {
	t.S.ClickEvent = nil
	t.S.HoverEvent = nil
	t.S.Insertion = nil
}

func filterTextChildren(children []c.Component) []c.Component {
	var out []c.Component
	for _, ch := range children {
		txt, ok := ch.(*c.Text)
		if !ok {
			continue
		}
		stripInteractableText(txt)
		txt.Extra = filterTextChildren(txt.Extra)
		out = append(out, txt)
	}
	return out
}
