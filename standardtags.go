package micromessage

import (
	"strings"

	c "go.minekube.com/common/minecraft/component"
)

// StandardTags exposes the built-in tag vocabulary as individually
// selectable TagResolvers, matching Adventure's StandardTags class. Combine
// them with TagResolverBuilder (or pass several to WithTags) to enable only
// a subset -- e.g. WithTags(StandardTags.Color()) allows "<red>" and
// "<color:...>" but leaves "<bold>" (and everything else) as literal text.
// StandardTags.All() is the default set used when WithTags is never called.
var StandardTags standardTags

type standardTags struct{}

// All returns every standard tag category combined -- the default tag set.
func (standardTags) All() TagResolver {
	return NewTagResolverBuilder().
		Resolver(StandardTags.Color()).
		Resolver(StandardTags.Decorations()).
		Resolver(StandardTags.Click()).
		Resolver(StandardTags.HoverEvent()).
		Resolver(StandardTags.Insertion()).
		Resolver(StandardTags.Font()).
		Resolver(StandardTags.Shadow()).
		Resolver(StandardTags.Gradient()).
		Resolver(StandardTags.Rainbow()).
		Resolver(StandardTags.Transition()).
		Resolver(StandardTags.Translatable()).
		Resolver(StandardTags.Newline()).
		Resolver(StandardTags.Reset()).
		Build()
}

// Color handles bare color tags ("<red>", "<#rrggbb>") and the explicit
// "<color:NAME>"/"<colour:NAME>"/"<c:NAME>" form.
func (standardTags) Color() TagResolver {
	return TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		n := strings.TrimPrefix(name, "!")
		switch n {
		case "color", "colour", "c":
			if a, ok := args.Pop(); ok {
				if col, ok := resolveColor(a); ok {
					return StylingTag(ColorStyle(col)), true
				}
			}
			return StylingTag(), true
		}
		if col, ok := resolveColor(n); ok {
			return StylingTag(ColorStyle(col)), true
		}
		return Tag{}, false
	})
}

// Decorations handles <bold>/<b>, <italic>/<i>/<em>, <underlined>/<u>,
// <strikethrough>/<st>, <obfuscated>/<obf>, their "<!name>" negations, and
// the explicit "<name:false>" boolean form.
func (standardTags) Decorations() TagResolver {
	return TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		negate := strings.HasPrefix(name, "!")
		n := strings.TrimPrefix(name, "!")
		dec, ok := decorationTags[n]
		if !ok {
			return Tag{}, false
		}
		value := !negate
		if a, ok := args.Pop(); ok {
			// "<!bold:true>" mixes the negation shorthand with the explicit
			// boolean form -- ambiguous, so the whole tag is left unresolved
			// (literal text), matching real MiniMessage.
			if negate {
				return Tag{}, false
			}
			switch strings.ToLower(a) {
			case "true":
				value = true
			case "false":
				value = false
			}
		}
		return StylingTag(DecorationStyle(dec, value)), true
	})
}

// Shadow handles "<shadow:NAME_OR_HEX:[alpha]>" and its "<!shadow>" negation.
func (standardTags) Shadow() TagResolver {
	return TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		negate := strings.HasPrefix(name, "!")
		if strings.TrimPrefix(name, "!") != "shadow" {
			return Tag{}, false
		}
		if negate {
			return StylingTag(ShadowStyle(&c.ShadowColor{})), true
		}
		if sc, ok := parseShadowColor(args.Rest()); ok {
			return StylingTag(ShadowStyle(sc)), true
		}
		return StylingTag(), true
	})
}

// Insertion handles "<insert:VALUE>".
func (standardTags) Insertion() TagResolver {
	return TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		if strings.TrimPrefix(name, "!") != "insert" {
			return Tag{}, false
		}
		if rest := args.Rest(); len(rest) > 0 {
			return StylingTag(InsertionStyle(strings.Join(rest, ":"))), true
		}
		return StylingTag(), true
	})
}

// Font handles "<font:KEY>" and "<font:NAMESPACE:KEY>".
func (standardTags) Font() TagResolver {
	return TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		if strings.TrimPrefix(name, "!") != "font" {
			return Tag{}, false
		}
		if k := parseKeyArgs(args.Rest()); k != nil {
			return StylingTag(FontStyle(k)), true
		}
		return StylingTag(), true
	})
}

// Click handles "<click:ACTION:VALUE>".
func (standardTags) Click() TagResolver {
	return TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		if strings.TrimPrefix(name, "!") != "click" {
			return Tag{}, false
		}
		rest := args.Rest()
		if len(rest) == 0 {
			return StylingTag(), true
		}
		// An unrecognized action (e.g. "<click:pet_a_kitty:...>") leaves the
		// whole tag unresolved, matching real MiniMessage -- unlike a
		// recognized action with a missing/odd value, which still matches.
		action, ok := clickAction(rest[0])
		if !ok {
			return Tag{}, false
		}
		value := ""
		if len(rest) > 1 {
			value = strings.Join(rest[1:], ":")
		}
		return StylingTag(ClickStyle(c.NewClickEvent(action, value))), true
	})
}

// HoverEvent handles "<hover:show_text:...>", "<hover:show_item:...>", and
// "<hover:show_entity:...>". Named HoverEvent (not Hover) to match
// Adventure's StandardTags.hoverEvent().
func (standardTags) HoverEvent() TagResolver {
	return TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		if strings.TrimPrefix(name, "!") != "hover" {
			return Tag{}, false
		}
		return Tag{kind: tagHover, hoverArgs: args.Rest()}, true
	})
}

// Translatable handles "<lang:KEY[:with...]>" ("tr", "translate") and the
// "<..._or:KEY:FALLBACK[:with...]>" fallback-carrying variants.
func (standardTags) Translatable() TagResolver {
	return TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		switch name {
		case "lang", "tr", "translate", "lang_or", "tr_or", "translate_or":
			// Every "with" argument is already carried as a tag argument, so
			// this tag never has (and never renders) children -- self-closing,
			// matching real MiniMessage, so that trailing content after an
			// unclosed "<lang:...>" stays as ordinary sibling text instead of
			// being swallowed as (unrendered) children.
			return Tag{kind: tagTranslatable, selfClosing: true}, true
		}
		return Tag{}, false
	})
}

// Newline handles "<br>"/"<newline>".
func (standardTags) Newline() TagResolver {
	return TagResolverFunc(func(name string, _ *ArgumentQueue) (Tag, bool) {
		if name == "br" || name == "newline" {
			return Tag{kind: tagText, value: "\n", selfClosing: true}, true
		}
		return Tag{}, false
	})
}

// Reset handles "<reset>" as a ParserDirective.
func (standardTags) Reset() TagResolver {
	return Resolver("reset", Reset)
}

// Gradient handles "<gradient[:c1:c2:...:cN][:phase]>".
func (standardTags) Gradient() TagResolver {
	return TagResolverFunc(func(name string, _ *ArgumentQueue) (Tag, bool) {
		if name != "gradient" {
			return Tag{}, false
		}
		return Tag{kind: tagGradient}, true
	})
}

// Rainbow handles "<rainbow[:phase][:!]>".
func (standardTags) Rainbow() TagResolver {
	return TagResolverFunc(func(name string, _ *ArgumentQueue) (Tag, bool) {
		if name != "rainbow" {
			return Tag{}, false
		}
		return Tag{kind: tagRainbow}, true
	})
}

// Transition handles "<transition:c1:c2:...:cN[:phase]>".
func (standardTags) Transition() TagResolver {
	return TagResolverFunc(func(name string, _ *ArgumentQueue) (Tag, bool) {
		if name != "transition" {
			return Tag{}, false
		}
		return Tag{kind: tagTransition}, true
	})
}
