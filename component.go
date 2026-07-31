package micromessage

import (
	"github.com/lucasb-eyer/go-colorful"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/key"
	mcolor "go.minekube.com/common/minecraft/color"
)

// TriState represents a decoration's tristate value: unset, true, or false.
type TriState int

const (
	Unset TriState = iota
	True
	False
)

// Decoration names
const (
	Bold          = "bold"
	Italic        = "italic"
	Underlined    = "underlined"
	Strikethrough = "strikethrough"
	Obfuscated    = "obfuscated"
)

var AllDecorations = []string{Bold, Italic, Underlined, Strikethrough, Obfuscated}

type ClickAction string

// Click actions
const (
	OpenURL         ClickAction = "open_url"
	OpenFile        ClickAction = "open_file"
	RunCommand      ClickAction = "run_command"
	SuggestCommand  ClickAction = "suggest_command"
	ChangePage      ClickAction = "change_page"
	CopyToClipboard ClickAction = "copy_to_clipboard"
)

type ClickEvent struct {
	Action ClickAction
	Value  string
}

type HoverAction string

// Hover actions
const (
	ShowText HoverAction = "show_text"
)

type HoverEvent struct {
	Action HoverAction
	Value  *Component
}

type Style struct {
	Color       *colorful.Color
	Decorations map[string]TriState
	ClickEvent  *ClickEvent
	HoverEvent  *HoverEvent
	Insertion   *string
	Font        *string
}

func NewStyle() Style {
	return Style{Decorations: map[string]TriState{}}
}

func (s Style) Clone() Style {
	c := Style{Decorations: map[string]TriState{}}
	if s.Color != nil {
		col := *s.Color
		c.Color = &col
	}
	if s.Decorations != nil {
		for k, v := range s.Decorations {
			c.Decorations[k] = v
		}
	}
	if s.ClickEvent != nil {
		ce := *s.ClickEvent
		c.ClickEvent = &ce
	}
	if s.HoverEvent != nil {
		he := *s.HoverEvent
		c.HoverEvent = &he
	}
	if s.Insertion != nil {
		ins := *s.Insertion
		c.Insertion = &ins
	}
	if s.Font != nil {
		f := *s.Font
		c.Font = &f
	}
	return c
}

func (s Style) Merge(other Style) Style {
	out := s.Clone()
	if other.Color != nil {
		col := *other.Color
		out.Color = &col
	}
	for k, v := range other.Decorations {
		out.Decorations[k] = v
	}
	if other.ClickEvent != nil {
		out.ClickEvent = other.ClickEvent
	}
	if other.HoverEvent != nil {
		out.HoverEvent = other.HoverEvent
	}
	if other.Insertion != nil {
		out.Insertion = other.Insertion
	}
	if other.Font != nil {
		out.Font = other.Font
	}
	return out
}

func (s Style) IsEmpty() bool {
	if s.Color != nil || s.ClickEvent != nil || s.HoverEvent != nil || s.Insertion != nil || s.Font != nil {
		return false
	}
	for _, v := range s.Decorations {
		if v != Unset {
			return false
		}
	}
	return true
}

type Component struct {
	Text     string
	Children []*Component
	Style    Style
}

func Text(s string) *Component {
	return &Component{Text: s, Style: NewStyle()}
}

func Empty() *Component {
	return &Component{Style: NewStyle()}
}

func (c *Component) Append(children ...*Component) *Component {
	c.Children = append(c.Children, children...)
	return c
}

func (c *Component) Clone() *Component {
	if c == nil {
		return nil
	}
	n := &Component{Text: c.Text, Style: c.Style.Clone()}
	for _, ch := range c.Children {
		n.Children = append(n.Children, ch.Clone())
	}
	return n
}

// ---------------------------------------------------------------------------
// Conversion to/from minekube's component API.
//
// The internal builder (buildComponent, resolveTagStyle, etc.) works with
// minekube's *component.Text / component.Style. These helpers convert between
// the public struct types and minekube's types at the Deserialize boundary.
// ---------------------------------------------------------------------------

// triToState converts our TriState to minekube's component.State.
func triToState(t TriState) component.State {
	switch t {
	case True:
		return component.True
	case False:
		return component.False
	default:
		return component.NotSet
	}
}

// stateToTri converts minekube's component.State to our TriState.
func stateToTri(s component.State) TriState {
	switch s {
	case component.True:
		return True
	case component.False:
		return False
	default:
		return Unset
	}
}

// styleToMinekube converts our public Style to minekube's component.Style.
func styleToMinekube(s Style) component.Style {
	ms := component.Style{}
	if s.Color != nil {
		rgb := (*mcolor.RGB)(s.Color)
		ms.Color = rgb
	}
	if v, ok := s.Decorations[Bold]; ok {
		ms.Bold = triToState(v)
	}
	if v, ok := s.Decorations[Italic]; ok {
		ms.Italic = triToState(v)
	}
	if v, ok := s.Decorations[Underlined]; ok {
		ms.Underlined = triToState(v)
	}
	if v, ok := s.Decorations[Strikethrough]; ok {
		ms.Strikethrough = triToState(v)
	}
	if v, ok := s.Decorations[Obfuscated]; ok {
		ms.Obfuscated = triToState(v)
	}
	if s.ClickEvent != nil {
		ms.ClickEvent = component.NewClickEvent(
			component.ClickActions[string(s.ClickEvent.Action)],
			s.ClickEvent.Value,
		)
	}
	if s.HoverEvent != nil && s.HoverEvent.Value != nil {
		ms.HoverEvent = component.ShowText(componentToMinekube(s.HoverEvent.Value))
	}
	if s.Insertion != nil {
		ms.Insertion = s.Insertion
	}
	if s.Font != nil {
		if k, err := key.Parse(*s.Font); err == nil {
			ms.Font = k
		}
	}
	return ms
}

// styleFromMinekube converts minekube's component.Style to our public Style.
func styleFromMinekube(ms component.Style) Style {
	s := NewStyle()
	if ms.Color != nil {
		r, g, b, _ := ms.Color.RGBA()
		col := colorful.Color{
			R: float64(r) / 65535.0,
			G: float64(g) / 65535.0,
			B: float64(b) / 65535.0,
		}
		s.Color = &col
	}
	s.Decorations[Bold] = stateToTri(ms.Bold)
	s.Decorations[Italic] = stateToTri(ms.Italic)
	s.Decorations[Underlined] = stateToTri(ms.Underlined)
	s.Decorations[Strikethrough] = stateToTri(ms.Strikethrough)
	s.Decorations[Obfuscated] = stateToTri(ms.Obfuscated)
	if ms.ClickEvent != nil {
		s.ClickEvent = &ClickEvent{
			Action: ClickAction(ms.ClickEvent.Action().Name()),
			Value:  ms.ClickEvent.Value(),
		}
	}
	if ms.HoverEvent != nil {
		if _, ok := ms.HoverEvent.Value().(component.Component); ok {
			s.HoverEvent = &HoverEvent{
				Action: ShowText,
				Value:  componentFromMinekube(ms.HoverEvent.Value().(component.Component)),
			}
		}
	}
	if ms.Insertion != nil {
		s.Insertion = ms.Insertion
	}
	if ms.Font != nil {
		f := ms.Font.String()
		s.Font = &f
	}
	return s
}

// componentToMinekube converts our public *Component tree to minekube's
// *component.Text.
func componentToMinekube(c *Component) *component.Text {
	if c == nil {
		return &component.Text{}
	}
	t := &component.Text{
		Content: c.Text,
		S:       styleToMinekube(c.Style),
	}
	for _, ch := range c.Children {
		t.Extra = append(t.Extra, componentToMinekube(ch))
	}
	return t
}

// componentFromMinekube converts minekube's component.Component to our public
// *Component tree.
func componentFromMinekube(mc component.Component) *Component {
	if mc == nil {
		return Empty()
	}
	t, ok := mc.(*component.Text)
	if !ok {
		return Empty()
	}
	c := &Component{
		Text:  t.Content,
		Style: styleFromMinekube(t.S),
	}
	for _, ch := range t.Extra {
		c.Children = append(c.Children, componentFromMinekube(ch))
	}
	return c
}

// ToMinekube converts a public *Component tree to minekube's component.Component
// interface, for interoperability with other go.minekube.com libraries (e.g. Gate).
func ToMinekube(c *Component) component.Component {
	return componentToMinekube(c)
}

// FromMinekube converts a minekube component.Component back to our public
// *Component tree.
func FromMinekube(mc component.Component) *Component {
	return componentFromMinekube(mc)
}

// ---------------------------------------------------------------------------
// minekube-side helpers used by the internal builder.
// minekube has no Clone/Merge methods, so we provide them here.
// ---------------------------------------------------------------------------

// cloneStyle clones a minekube component.Style.
func cloneStyle(s component.Style) component.Style {
	return component.Style{
		Obfuscated:    s.Obfuscated,
		Bold:          s.Bold,
		Strikethrough: s.Strikethrough,
		Underlined:    s.Underlined,
		Italic:        s.Italic,
		Font:          s.Font,
		Color:         s.Color,
		ClickEvent:    s.ClickEvent,
		HoverEvent:    s.HoverEvent,
		Insertion:     s.Insertion,
		ShadowColor:   s.ShadowColor,
	}
}

// mergeStyle returns a clone of parent with non-zero fields from other overlaid.
func mergeStyle(parent, other component.Style) component.Style {
	out := cloneStyle(parent)
	if other.Color != nil {
		out.Color = other.Color
	}
	if other.Bold != component.NotSet {
		out.Bold = other.Bold
	}
	if other.Italic != component.NotSet {
		out.Italic = other.Italic
	}
	if other.Underlined != component.NotSet {
		out.Underlined = other.Underlined
	}
	if other.Strikethrough != component.NotSet {
		out.Strikethrough = other.Strikethrough
	}
	if other.Obfuscated != component.NotSet {
		out.Obfuscated = other.Obfuscated
	}
	if other.ClickEvent != nil {
		out.ClickEvent = other.ClickEvent
	}
	if other.HoverEvent != nil {
		out.HoverEvent = other.HoverEvent
	}
	if other.Insertion != nil {
		out.Insertion = other.Insertion
	}
	if other.Font != nil {
		out.Font = other.Font
	}
	if other.ShadowColor != nil {
		out.ShadowColor = other.ShadowColor
	}
	return out
}

// cloneText deep-clones a minekube *component.Text.
func cloneText(t *component.Text) *component.Text {
	if t == nil {
		return &component.Text{}
	}
	cp := &component.Text{
		Content: t.Content,
		S:       cloneStyle(t.S),
	}
	for _, ch := range t.Extra {
		cp.Extra = append(cp.Extra, cloneText(asText(ch)))
	}
	return cp
}

// asText narrows a component.Component to *component.Text.
func asText(c component.Component) *component.Text {
	if c == nil {
		return &component.Text{}
	}
	if t, ok := c.(*component.Text); ok {
		return t
	}
	return &component.Text{Extra: c.Children()}
}