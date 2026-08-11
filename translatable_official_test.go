package micromessage

import (
	"testing"

	c "go.minekube.com/common/minecraft/component"
)

// trOfficialColor returns the hex color of a style, "" if unset.
func trOfficialColor(s c.Style) string {
	if s.Color != nil {
		return s.Color.Hex()
	}
	return ""
}

// trOfficialRoot deserializes src and returns its top-level children.
func trOfficialRoot(t *testing.T, src string) []c.Component {
	t.Helper()
	root, err := Deserialize(src)
	if err != nil {
		t.Fatalf("deserialize error for %q: %v", src, err)
	}
	return root.Children()
}

func trOfficialText(t *testing.T, comp c.Component) *c.Text {
	t.Helper()
	txt, ok := comp.(*c.Text)
	if !ok {
		t.Fatalf("expected *c.Text, got %T", comp)
	}
	return txt
}

func trOfficialTranslation(t *testing.T, comp c.Component) *c.Translation {
	t.Helper()
	tr, ok := comp.(*c.Translation)
	if !ok {
		t.Fatalf("expected *c.Translation, got %T", comp)
	}
	return tr
}

// Ported from TranslatableTagTest#testTranslatable.
func TestTranslatableOfficial_Basic(t *testing.T) {
	children := trOfficialRoot(t, "You should get a <lang:block.minecraft.diamond_block>!")
	if len(children) != 3 {
		t.Fatalf("got %d children, want 3: %#v", len(children), children)
	}
	lead := trOfficialText(t, children[0])
	if lead.Content != "You should get a " {
		t.Fatalf("lead content = %q", lead.Content)
	}
	tr := trOfficialTranslation(t, children[1])
	if tr.Key != "block.minecraft.diamond_block" {
		t.Fatalf("key = %q", tr.Key)
	}
	if tr.Fallback != "" {
		t.Fatalf("fallback = %q, want empty", tr.Fallback)
	}
	if len(tr.With) != 0 {
		t.Fatalf("with = %#v, want empty", tr.With)
	}
	tail := trOfficialText(t, children[2])
	if tail.Content != "!" {
		t.Fatalf("tail content = %q", tail.Content)
	}
}

// Ported from TranslatableTagTest#testTranslatableWith.
func TestTranslatableOfficial_With(t *testing.T) {
	children := trOfficialRoot(t, "Test: <lang:commands.drop.success.single:'<red>1':'<blue>Stone'>!")
	if len(children) != 3 {
		t.Fatalf("got %d children, want 3: %#v", len(children), children)
	}
	lead := trOfficialText(t, children[0])
	if lead.Content != "Test: " {
		t.Fatalf("lead content = %q", lead.Content)
	}
	tr := trOfficialTranslation(t, children[1])
	if tr.Key != "commands.drop.success.single" {
		t.Fatalf("key = %q", tr.Key)
	}
	if len(tr.With) != 2 {
		t.Fatalf("with = %#v, want 2 entries", tr.With)
	}
	w0 := trOfficialText(t, tr.With[0])
	if w0.Content != "1" || trOfficialColor(w0.S) != "#ff5555" {
		t.Fatalf("with[0] = %q color=%s", w0.Content, trOfficialColor(w0.S))
	}
	w1 := trOfficialText(t, tr.With[1])
	if w1.Content != "Stone" || trOfficialColor(w1.S) != "#5555ff" {
		t.Fatalf("with[1] = %q color=%s", w1.Content, trOfficialColor(w1.S))
	}
	tail := trOfficialText(t, children[2])
	if tail.Content != "!" {
		t.Fatalf("tail content = %q", tail.Content)
	}
}

// Ported from TranslatableTagTest#testTranslatableWithHover.
func TestTranslatableOfficial_WithHover(t *testing.T) {
	children := trOfficialRoot(t, `Test: <lang:commands.drop.success.single:'<hover:show_text:\'<red>dum\'><red>1':'<blue>Stone'>!`)
	if len(children) != 3 {
		t.Fatalf("got %d children, want 3: %#v", len(children), children)
	}
	tr := trOfficialTranslation(t, children[1])
	if len(tr.With) != 2 {
		t.Fatalf("with = %#v, want 2 entries", tr.With)
	}
	w0 := trOfficialText(t, tr.With[0])
	if w0.Content != "1" || trOfficialColor(w0.S) != "#ff5555" {
		t.Fatalf("with[0] = %q color=%s", w0.Content, trOfficialColor(w0.S))
	}
	if w0.S.HoverEvent == nil {
		t.Fatalf("with[0] missing hover event")
	}
	if w0.S.HoverEvent.Action().Name() != "show_text" {
		t.Fatalf("with[0] hover action = %s", w0.S.HoverEvent.Action().Name())
	}
	hoverText := trOfficialText(t, w0.S.HoverEvent.Value().(c.Component))
	if hoverText.Content != "dum" || trOfficialColor(hoverText.S) != "#ff5555" {
		t.Fatalf("with[0] hover text = %q color=%s", hoverText.Content, trOfficialColor(hoverText.S))
	}
	w1 := trOfficialText(t, tr.With[1])
	if w1.Content != "Stone" || trOfficialColor(w1.S) != "#5555ff" {
		t.Fatalf("with[1] = %q color=%s", w1.Content, trOfficialColor(w1.S))
	}
	tail := trOfficialText(t, children[2])
	if tail.Content != "!" {
		t.Fatalf("tail content = %q", tail.Content)
	}
}

// Ported from TranslatableTagTest#testKingAlter.
func TestTranslatableOfficial_KingAlter(t *testing.T) {
	children := trOfficialRoot(t, "Ahoy <lang:offset.-40:'<red>mates!'>")
	if len(children) != 2 {
		t.Fatalf("got %d children, want 2: %#v", len(children), children)
	}
	lead := trOfficialText(t, children[0])
	if lead.Content != "Ahoy " {
		t.Fatalf("lead content = %q", lead.Content)
	}
	tr := trOfficialTranslation(t, children[1])
	if tr.Key != "offset.-40" {
		t.Fatalf("key = %q", tr.Key)
	}
	if len(tr.With) != 1 {
		t.Fatalf("with = %#v, want 1 entry", tr.With)
	}
	w0 := trOfficialText(t, tr.With[0])
	if w0.Content != "mates!" || trOfficialColor(w0.S) != "#ff5555" {
		t.Fatalf("with[0] = %q color=%s", w0.Content, trOfficialColor(w0.S))
	}
}

// Ported from TranslatableFallbackTagTest#testTranslatable.
func TestTranslatableFallbackOfficial_Basic(t *testing.T) {
	children := trOfficialRoot(t, "You should get a <lang_or:block.minecraft.diamond_block:'Diamond Block'>!")
	if len(children) != 3 {
		t.Fatalf("got %d children, want 3: %#v", len(children), children)
	}
	lead := trOfficialText(t, children[0])
	if lead.Content != "You should get a " {
		t.Fatalf("lead content = %q", lead.Content)
	}
	tr := trOfficialTranslation(t, children[1])
	if tr.Key != "block.minecraft.diamond_block" {
		t.Fatalf("key = %q", tr.Key)
	}
	if tr.Fallback != "Diamond Block" {
		t.Fatalf("fallback = %q", tr.Fallback)
	}
	if len(tr.With) != 0 {
		t.Fatalf("with = %#v, want empty", tr.With)
	}
	tail := trOfficialText(t, children[2])
	if tail.Content != "!" {
		t.Fatalf("tail content = %q", tail.Content)
	}
}

// tr/translate are aliases for lang; lang_or/tr_or/translate_or share the
// fallback-carrying variant. Spot-check the aliases resolve identically.
func TestTranslatableOfficial_Aliases(t *testing.T) {
	for _, name := range []string{"lang", "tr", "translate"} {
		children := trOfficialRoot(t, "<"+name+":some.key:'<red>x'>")
		if len(children) != 1 {
			t.Fatalf("%s: got %d children, want 1: %#v", name, len(children), children)
		}
		tr := trOfficialTranslation(t, children[0])
		if tr.Key != "some.key" || len(tr.With) != 1 {
			t.Fatalf("%s: key=%q with=%#v", name, tr.Key, tr.With)
		}
	}
	for _, name := range []string{"lang_or", "tr_or", "translate_or"} {
		children := trOfficialRoot(t, "<"+name+":some.key:'fb':'<red>x'>")
		if len(children) != 1 {
			t.Fatalf("%s: got %d children, want 1: %#v", name, len(children), children)
		}
		tr := trOfficialTranslation(t, children[0])
		if tr.Key != "some.key" || tr.Fallback != "fb" || len(tr.With) != 1 {
			t.Fatalf("%s: key=%q fallback=%q with=%#v", name, tr.Key, tr.Fallback, tr.With)
		}
	}
}
