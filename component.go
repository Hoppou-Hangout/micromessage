package micromessage

import "github.com/lucasb-eyer/go-colorful"

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
	c := Style{}
	if s.Color != nil {
		col := *s.Color
		c.Color = &col
	}
	if s.Decorations != nil {
		c.Decorations = make(map[string]TriState, len(s.Decorations))
		for k, v := range s.Decorations {
			c.Decorations[k] = v
		}
	}
	c.ClickEvent = s.ClickEvent
	c.HoverEvent = s.HoverEvent
	c.Insertion = s.Insertion
	c.Font = s.Font
	return c
}

func (s Style) Merge(other Style) Style {
	out := s.Clone()
	if other.Color != nil {
		col := *other.Color
		out.Color = &col
	}
	for k, v := range other.Decorations {
		if out.Decorations == nil {
			out.Decorations = map[string]TriState{}
		}
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
