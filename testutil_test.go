package micromessage

import (
	"fmt"
	"testing"

	c "go.minekube.com/common/minecraft/component"
)

// charLeaf is one rendered character plus the style attributes that apply to
// it, flattened out of the component.Component tree so we can compare directly
// against Adventure's own per-character expected values.
type charLeaf struct {
	ch            rune
	color         string // hex, "" if unset
	bold          bool
	italic        bool
	underlined    bool
	strikethrough bool
	obfuscated    bool
	click         string // "action:value", empty if none
	hover         string
}

func renderSrc(t *testing.T, src string) []charLeaf {
	t.Helper()
	root, err := Deserialize(src)
	if err != nil {
		t.Fatalf("deserialize error for %q: %v", src, err)
	}
	return flattenComponent(root)
}

func flattenComponent(comp c.Component) []charLeaf {
	var out []charLeaf
	if txt, ok := comp.(*c.Text); ok && txt.Content != "" {
		s := txt.S
		var l charLeaf
		if s.Color != nil {
			l.color = s.Color.Hex()
		}
		l.bold = s.Bold == c.True
		l.italic = s.Italic == c.True
		l.underlined = s.Underlined == c.True
		l.strikethrough = s.Strikethrough == c.True
		l.obfuscated = s.Obfuscated == c.True
		if s.ClickEvent != nil {
			l.click = s.ClickEvent.Action().Name() + ":" + s.ClickEvent.Value()
		}
		if s.HoverEvent != nil {
			if txt2, ok := s.HoverEvent.Value().(*c.Text); ok {
				l.hover = s.HoverEvent.Action().Name() + ":" + txt2.Content
			}
		}
		for _, r := range txt.Content {
			cp := l
			cp.ch = r
			out = append(out, cp)
		}
	}
	for _, e := range comp.Children() {
		out = append(out, flattenComponent(e)...)
	}
	return out
}

func text(s, color string) []charLeaf {
	var out []charLeaf
	for _, r := range s {
		out = append(out, charLeaf{ch: r, color: color})
	}
	return out
}

func textB(s, color string) []charLeaf {
	l := text(s, color)
	for i := range l {
		l[i].bold = true
	}
	return l
}

func grad(ch byte, hexes ...string) []charLeaf {
	out := make([]charLeaf, len(hexes))
	for i, h := range hexes {
		out[i] = charLeaf{ch: rune(ch), color: h}
	}
	return out
}

func gradRunes(runes string, hexes ...string) []charLeaf {
	rs := []rune(runes)
	out := make([]charLeaf, len(hexes))
	for i, h := range hexes {
		out[i] = charLeaf{ch: rs[i], color: h}
	}
	return out
}

func cat(groups ...[]charLeaf) []charLeaf {
	var out []charLeaf
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

func check(t *testing.T, name, src string, want []charLeaf) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		got := renderSrc(t, src)
		if len(got) != len(want) {
			t.Fatalf("%s: got %d chars, want %d\n got: %s\nwant: %s", src, len(got), len(want), fmtLeaves(got), fmtLeaves(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s: char %d mismatch\n got: %+v\nwant: %+v\nfull got:  %s\nfull want: %s",
					src, i, got[i], want[i], fmtLeaves(got), fmtLeaves(want))
			}
		}
	})
}

func fmtLeaves(ls []charLeaf) string {
	s := ""
	for _, l := range ls {
		extra := ""
		if l.bold {
			extra += " bold"
		}
		if l.click != "" {
			extra += " click=" + l.click
		}
		if l.hover != "" {
			extra += " hover=" + l.hover
		}
		s += fmt.Sprintf("[%c %s%s]", l.ch, l.color, extra)
	}
	return s
}
