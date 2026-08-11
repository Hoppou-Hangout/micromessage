package micromessage

import (
	"fmt"
	"strings"

	mccolor "go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/key"
)

// --- Tag -------------------------------------------------------------------

type tagKind int

const (
	tagText tagKind = iota
	tagParsed
	tagComponent
	tagStyling
	tagModifying
	tagDirective
	// The remaining kinds back the built-in structural tags (StandardTags),
	// which need the whole node (for children/args) rather than just a
	// resolved value -- render() special-cases them using the existing node
	// it already has in hand.
	tagHover
	tagTranslatable
	tagGradient
	tagRainbow
	tagTransition
)

// Tag is what a TagResolver produces for a tag it recognizes. Build one with
// Text, Parsed, ComponentTag, StylingTag, or ModifyingTag -- matching
// Adventure's three Tag kinds (PreProcess, Inserting, Modifying) plus
// ParserDirective for tags like RESET.
type Tag struct {
	kind        tagKind
	value       string
	selfClosing bool
	comp        c.Component
	styles      []StyleApplicable
	modifying   ModifyingTag
	directive   ParserDirective
	hoverArgs   []string
}

// Text returns a Tag whose value is inserted as literal text, taking on the
// ambient style at the tag's position but never itself re-parsed as
// MiniMessage. It is self-closing by default, matching Adventure's
// Placeholder.unparsed.
func Text(value string) Tag { return Tag{kind: tagText, value: value, selfClosing: true} }

// Parsed returns a Tag whose value is parsed as its own MiniMessage snippet
// and spliced in at the tag's position, inheriting the ambient style. The
// same TagResolvers apply recursively inside it (capped at maxTagDepth to
// catch self-reference). This is a pragmatic stand-in for Adventure's
// PreProcess tags: rather than splicing the raw string back into the input
// and re-lexing (which needs a resolver-aware, single-pass parser), it is
// parsed and rendered as an independent sub-document at render time. For all
// but pathological cases (tag boundaries split across the substitution) the
// result is the same. Self-closing by default -- it never wraps content of
// its own, so trailing input after it stays ordinary sibling text.
func Parsed(value string) Tag { return Tag{kind: tagParsed, value: value, selfClosing: true} }

// ComponentTag returns a Tag that inserts comp verbatim, with whatever style
// comp already carries -- the ambient style at the tag's position is not
// applied to it. Self-closing by default, matching Adventure's
// Placeholder.component.
func ComponentTag(comp c.Component) Tag {
	return Tag{kind: tagComponent, comp: comp, selfClosing: true}
}

// StylingTag returns an Inserting Tag that applies a set of style changes to
// its wrapped content, matching Adventure's Tag.styling(StyleBuilderApplicable...).
// Not self-closing: it wraps whatever follows until a matching close tag (or
// EOF), same as the built-in <color>/<bold>/<click>/... tags.
func StylingTag(styles ...StyleApplicable) Tag { return Tag{kind: tagStyling, styles: styles} }

// ModifyingTagValue returns a Tag backed by a custom ModifyingTag
// implementation, matching Adventure's Modifying tags (used internally for
// <gradient> and <rainbow>).
func ModifyingTagValue(m ModifyingTag) Tag { return Tag{kind: tagModifying, modifying: m} }

// DirectiveTag returns a Tag that behaves as a parser directive, e.g. an
// alternate spelling of <reset>. See ParserDirective.
func DirectiveTag(d ParserDirective) Tag { return Tag{kind: tagDirective, directive: d} }

// SelfClosing marks an Inserting Tag (Text/ComponentTag/StylingTag) as never
// taking a close tag or wrapping content, even if the input writes it as
// <tag>...</tag> or leaves it unclosed. Text and ComponentTag are
// self-closing by default; use this to opt StylingTag/custom tags in too.
func (t Tag) SelfClosing() Tag {
	t.selfClosing = true
	return t
}

// --- Parser directives -------------------------------------------------

// ParserDirective is an instruction to the parser rather than something that
// produces or modifies a component, matching Adventure's ParserDirective.
type ParserDirective int

const (
	// Reset indicates that this tag should close all currently open tags,
	// exactly like <reset>. Registering a resolver for a custom tag name via
	// DirectiveTag(Reset) gives it identical behavior to <reset> under a
	// different name.
	Reset ParserDirective = iota
)

// --- Modifying tags ------------------------------------------------------

// ModifyingTag is the interface custom Modifying tags implement, matching
// Adventure's Modifying tag kind (used internally for <gradient> and
// <rainbow>). Visit is called once per node in the wrapped content, in
// depth-first order, before any component is produced; PostVisit is called
// once after the full traversal; then Apply is called once per produced
// child component, in order, to transform it.
//
// If a Modifying tag carries state across Visit calls, its TagResolver must
// return a fresh instance per tag occurrence -- state is not reset between
// uses.
type ModifyingTag interface {
	Visit(node *Node)
	PostVisit()
	Apply(current c.Component, depth int) c.Component
}

// --- Style applicables -----------------------------------------------------

// StyleApplicable is one change StylingTag applies to a c.Style, matching
// Adventure's StyleBuilderApplicable. Build one with ColorStyle,
// DecorationStyle, ClickStyle, HoverStyle, FontStyle, InsertionStyle, or
// ShadowStyle.
type StyleApplicable interface {
	applyStyle(c.Style) c.Style
}

type styleApplicableFunc func(c.Style) c.Style

func (f styleApplicableFunc) applyStyle(s c.Style) c.Style { return f(s) }

func ColorStyle(col mccolor.Color) StyleApplicable {
	return styleApplicableFunc(func(s c.Style) c.Style { s.Color = col; return s })
}

func DecorationStyle(dec c.Decoration, on bool) StyleApplicable {
	return styleApplicableFunc(func(s c.Style) c.Style { s.SetDecoration(dec, c.StateByBool(on)); return s })
}

func ClickStyle(ev c.ClickEvent) StyleApplicable {
	return styleApplicableFunc(func(s c.Style) c.Style { s.ClickEvent = ev; return s })
}

func HoverStyle(ev c.HoverEvent) StyleApplicable {
	return styleApplicableFunc(func(s c.Style) c.Style { s.HoverEvent = ev; return s })
}

func FontStyle(k key.Key) StyleApplicable {
	return styleApplicableFunc(func(s c.Style) c.Style { s.Font = k; return s })
}

func InsertionStyle(v string) StyleApplicable {
	return styleApplicableFunc(func(s c.Style) c.Style { s.Insertion = &v; return s })
}

func ShadowStyle(sc *c.ShadowColor) StyleApplicable {
	return styleApplicableFunc(func(s c.Style) c.Style { s.ShadowColor = sc; return s })
}

func applyStyles(base c.Style, styles []StyleApplicable) c.Style {
	for _, sa := range styles {
		base = sa.applyStyle(base)
	}
	return base
}

// --- Tag resolvers -----------------------------------------------------

// TagResolver supplies a Tag for a tag occurrence (name + arguments) the
// parser encounters. Built-in tags (color names, "bold", "gradient", ...)
// are themselves ordinary TagResolvers -- see StandardTags -- so the active
// tag vocabulary is entirely determined by which resolvers are in play; see
// WithTags and WithTagResolver.
//
// Tag names passed to ResolveTag are already lower-cased (a leading
// "!"/"?"/"#" sigil, if any, is preserved), matching Adventure's
// case-insensitive tag name resolution.
type TagResolver interface {
	ResolveTag(name string, args *ArgumentQueue) (tag Tag, ok bool)
}

// TagResolverFunc adapts a plain function to a TagResolver.
type TagResolverFunc func(name string, args *ArgumentQueue) (Tag, bool)

func (f TagResolverFunc) ResolveTag(name string, args *ArgumentQueue) (Tag, bool) {
	return f(name, args)
}

// Placeholder returns a TagResolver that matches a single tag name
// case-insensitively and always resolves to tag, ignoring any arguments --
// the common case (a named value substituted wherever <name> appears).
func Placeholder(name string, tag Tag) TagResolver {
	lower := strings.ToLower(name)
	return TagResolverFunc(func(n string, _ *ArgumentQueue) (Tag, bool) {
		if n == lower {
			return tag, true
		}
		return Tag{}, false
	})
}

// Resolver returns a TagResolver that matches a single tag name
// case-insensitively and resolves to a ParserDirective, e.g. a "<clear>" tag
// that behaves exactly like "<reset>".
func Resolver(name string, d ParserDirective) TagResolver {
	lower := strings.ToLower(name)
	return TagResolverFunc(func(n string, _ *ArgumentQueue) (Tag, bool) {
		if n == lower {
			return DirectiveTag(d), true
		}
		return Tag{}, false
	})
}

// TagResolverBuilder composes multiple TagResolvers into one, tried in the
// order added, matching Adventure's TagResolver.builder().
type TagResolverBuilder struct {
	resolvers []TagResolver
}

// NewTagResolverBuilder returns an empty TagResolverBuilder, matching
// Adventure's TagResolver.builder().
func NewTagResolverBuilder() *TagResolverBuilder { return &TagResolverBuilder{} }

// Resolver appends r to the set of resolvers this builder combines.
func (b *TagResolverBuilder) Resolver(r TagResolver) *TagResolverBuilder {
	b.resolvers = append(b.resolvers, r)
	return b
}

// Build returns a single TagResolver that tries the most-recently-added
// resolver first, matching Adventure's TagResolverBuilder (the last resolver
// added for a given name wins).
func (b *TagResolverBuilder) Build() TagResolver {
	resolvers := make([]TagResolver, len(b.resolvers))
	for i, r := range b.resolvers {
		resolvers[len(b.resolvers)-1-i] = r
	}
	return combinedResolver(resolvers)
}

type combinedResolver []TagResolver

func (rs combinedResolver) ResolveTag(name string, args *ArgumentQueue) (Tag, bool) {
	for _, r := range rs {
		if r == nil {
			continue
		}
		args.reset()
		if tag, ok := r.ResolveTag(name, args); ok {
			return tag, true
		}
	}
	return Tag{}, false
}

// resolveTag tries each resolver in order and returns the first match. name
// is lower-cased before matching, per TagResolver's contract.
func resolveTag(resolvers []TagResolver, name string, rawArgs []string) (Tag, bool) {
	lower := strings.ToLower(name)
	for _, r := range resolvers {
		if r == nil {
			continue
		}
		q := newArgumentQueue(rawArgs)
		if tag, ok := r.ResolveTag(lower, q); ok {
			return tag, true
		}
	}
	return Tag{}, false
}

// maxTagDepth bounds recursive expansion of Parsed tags, so a
// self-referential placeholder (e.g. Placeholder("foo", Parsed("<foo>")))
// fails fast instead of recursing until the stack overflows.
const maxTagDepth = 64

// --- ArgumentQueue -----------------------------------------------------

// ArgumentQueue gives a TagResolver access to a tag's raw arguments,
// matching Adventure's ArgumentQueue/Tag.Argument.
type ArgumentQueue struct {
	args []string
	pos  int
}

func newArgumentQueue(args []string) *ArgumentQueue { return &ArgumentQueue{args: args} }

func (q *ArgumentQueue) reset() { q.pos = 0 }

// HasNext reports whether another argument remains.
func (q *ArgumentQueue) HasNext() bool { return q.pos < len(q.args) }

// Remaining returns how many arguments are left.
func (q *ArgumentQueue) Remaining() int { return len(q.args) - q.pos }

// Pop returns the next argument and advances the queue, or ok=false if
// there isn't one.
func (q *ArgumentQueue) Pop() (value string, ok bool) {
	if !q.HasNext() {
		return "", false
	}
	v := q.args[q.pos]
	q.pos++
	return v, true
}

// PopOr returns the next argument and advances the queue, or panics with a
// *ArgumentError carrying errorMessage if there isn't one -- matching
// Adventure's Tag.Argument#popOr, which interrupts tag resolution on missing
// required arguments. Deserialize recovers this into a normal error.
func (q *ArgumentQueue) PopOr(errorMessage string) string {
	v, ok := q.Pop()
	if !ok {
		panic(&ArgumentError{errorMessage})
	}
	return v
}

// PeekOr is like PopOr but does not advance the queue.
func (q *ArgumentQueue) PeekOr(errorMessage string) string {
	if !q.HasNext() {
		panic(&ArgumentError{errorMessage})
	}
	return q.args[q.pos]
}

// Rest returns every remaining argument and drains the queue.
func (q *ArgumentQueue) Rest() []string {
	rest := q.args[q.pos:]
	q.pos = len(q.args)
	return rest
}

// ArgumentError is the panic value PopOr/PeekOr raise on a missing required
// argument; Deserialize recovers it into a plain error.
type ArgumentError struct{ Message string }

func (e *ArgumentError) Error() string { return e.Message }

// --- Preprocessing -------------------------------------------------------

// Preprocessor transforms the raw input string before it is lexed and
// parsed, e.g. to translate legacy '&'-formatted color codes into
// MiniMessage tags. Preprocessors run in the order passed to Deserialize,
// each seeing the previous one's output.
type Preprocessor func(input string) string

// --- Options -------------------------------------------------------------

// Option configures a Deserialize call.
type Option func(*options)

type options struct {
	tags          []TagResolver // replaces the standard tag set if non-nil
	extra         []TagResolver // appended after tags (or the standard set)
	preprocessors []Preprocessor
	postProcess   []func(*c.Text) *c.Text
	strict        bool
	debug         func(string)
}

// WithTags replaces the standard built-in tag set with resolver, matching
// Adventure's MiniMessage.builder().tags(resolver). Built-in tags not
// included in resolver (directly or via StandardTags.All()) stop being
// recognized and are rendered as literal text. If never called, the default
// is StandardTags.All().
func WithTags(resolver TagResolver) Option {
	return func(o *options) { o.tags = append(o.tags, resolver) }
}

// WithTagResolver registers an additional TagResolver, tried after the
// active tag set (the standard set, or whatever WithTags supplied). This is
// the common way to add placeholders/dynamic tags without disabling any
// built-ins, matching Adventure's builder.editTags(b -> b.resolver(...)).
func WithTagResolver(r TagResolver) Option {
	return func(o *options) { o.extra = append(o.extra, r) }
}

// WithPreprocessor registers a Preprocessor run over the raw input string
// before parsing. Multiple preprocessors may be passed (or this option
// repeated); they run in the order added, each seeing the previous one's
// output.
func WithPreprocessor(p Preprocessor) Option {
	return func(o *options) { o.preprocessors = append(o.preprocessors, p) }
}

// WithStrict enables strict mode: an unclosed tag is a parse error instead
// of being auto-closed at EOF, matching Adventure's Builder#strict(true).
// Tags that simply don't resolve to anything are still rendered as literal
// text either way.
func WithStrict(strict bool) Option {
	return func(o *options) { o.strict = strict }
}

// WithDebug registers a callback that receives diagnostic messages about why
// a tag failed to resolve, matching Adventure's Builder#debug(Consumer<String>).
func WithDebug(fn func(string)) Option {
	return func(o *options) { o.debug = fn }
}

func (o *options) log(format string, args ...any) {
	if o.debug != nil {
		o.debug(fmt.Sprintf(format, args...))
	}
}

// resolvers returns the final, ordered TagResolver chain for a Deserialize
// call: anything added via WithTagResolver, tried first so it can shadow a
// built-in of the same name, followed by the active tag set (WithTags, or
// StandardTags.All() by default) -- matching Adventure, where a per-call
// resolver takes priority over the base tag set.
func (o *options) resolvers() []TagResolver {
	base := o.tags
	if base == nil {
		base = []TagResolver{StandardTags.All()}
	}
	extra := make([]TagResolver, len(o.extra))
	for i, r := range o.extra {
		extra[len(o.extra)-1-i] = r // last-added WithTagResolver wins, like TagResolverBuilder
	}
	return append(extra, base...)
}
