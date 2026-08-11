package micromessage

import (
	"testing"

	"github.com/google/uuid"
	c "go.minekube.com/common/minecraft/component"
)

// Ported from adventure's ClickTagTest.java / HoverTagTest.java. Serialization
// (Component -> MiniMessage string) tests are skipped: this project only
// implements deserialization.

// interactiveOfficialPlainText concatenates the literal text content of a
// component tree (Text/Translation), ignoring styling, for hover-value
// assertions where the flat charLeaf.hover shortcut doesn't apply (e.g. the
// hover text itself is made of multiple styled/placeholder-substituted parts).
func interactiveOfficialPlainText(comp c.Component) string {
	var s string
	if txt, ok := comp.(*c.Text); ok {
		s += txt.Content
	}
	for _, e := range comp.Children() {
		s += interactiveOfficialPlainText(e)
	}
	return s
}

func TestOfficialClick(t *testing.T) {
	got := renderSrc(t, "<click:run_command:test>TEST")
	want := text("TEST", "")
	if len(got) != len(want) {
		t.Fatalf("got %d chars, want %d", len(got), len(want))
	}
	for i, l := range got {
		if l.ch != want[i].ch {
			t.Fatalf("char %d: got %q, want %q", i, l.ch, want[i].ch)
		}
		if l.click != "run_command:test" {
			t.Fatalf("char %d: click = %q, want run_command:test", i, l.click)
		}
	}
}

func TestOfficialClickExtendedCommand(t *testing.T) {
	got := renderSrc(t, "<click:run_command:/test command>TEST")
	want := text("TEST", "")
	if len(got) != len(want) {
		t.Fatalf("got %d chars, want %d", len(got), len(want))
	}
	for i, l := range got {
		if l.ch != want[i].ch {
			t.Fatalf("char %d: got %q, want %q", i, l.ch, want[i].ch)
		}
		if l.click != "run_command:/test command" {
			t.Fatalf("char %d: click = %q, want run_command:/test command", i, l.click)
		}
	}
}

func TestOfficialInvalidClick(t *testing.T) {
	check(t, "testInvalidClick",
		"<click:pet_a_kitty:'a very cute one'>best click event",
		text("<click:pet_a_kitty:'a very cute one'>best click event", ""))
}

func TestOfficialHover(t *testing.T) {
	got := renderSrc(t, `<hover:show_text:"<red>test">TEST`)
	want := text("TEST", "")
	if len(got) != len(want) {
		t.Fatalf("got %d chars, want %d", len(got), len(want))
	}
	for i, l := range got {
		if l.ch != want[i].ch {
			t.Fatalf("char %d: got %q, want %q", i, l.ch, want[i].ch)
		}
		if l.hover != "show_text:test" {
			t.Fatalf("char %d: hover = %q, want show_text:test", i, l.hover)
		}
	}
}

func TestOfficialHover2(t *testing.T) {
	got := renderSrc(t, `<hover:show_text:'<red>test'>TEST`)
	for i, l := range got {
		if l.hover != "show_text:test" {
			t.Fatalf("char %d: hover = %q, want show_text:test", i, l.hover)
		}
	}
}

func TestOfficialHoverWithColon(t *testing.T) {
	got := renderSrc(t, `<hover:show_text:"<red>test:TEST">TEST`)
	for i, l := range got {
		if l.hover != "show_text:test:TEST" {
			t.Fatalf("char %d: hover = %q, want show_text:test:TEST", i, l.hover)
		}
	}
}

func TestOfficialHoverMultiline(t *testing.T) {
	got := renderSrc(t, "<hover:show_text:'<red>test\ntest2'>TEST")
	for i, l := range got {
		if l.hover != "show_text:test\ntest2" {
			t.Fatalf("char %d: hover = %q, want show_text:test\\ntest2", i, l.hover)
		}
	}
}

func TestOfficialHoverWithInsertingComponent(t *testing.T) {
	root := deserialize(t, `<red><hover:show_text:"Test"><lang:item.minecraft.stick>`)
	if len(root.Extra) != 1 {
		t.Fatalf("want 1 extra component, got %d: %+v", len(root.Extra), root.Extra)
	}
	tr, ok := root.Extra[0].(*c.Translation)
	if !ok {
		t.Fatalf("extra[0] = %T, want *c.Translation", root.Extra[0])
	}
	if tr.Key != "item.minecraft.stick" {
		t.Fatalf("translation key = %q, want item.minecraft.stick", tr.Key)
	}
	if tr.S.Color == nil || tr.S.Color.Hex() != mustColor(t, "red").Hex() {
		t.Fatalf("translation color = %v, want red", tr.S.Color)
	}
	if tr.S.HoverEvent == nil || tr.S.HoverEvent.Action().Name() != "show_text" {
		t.Fatalf("hover event = %v, want show_text", tr.S.HoverEvent)
	}
	hoverText, ok := tr.S.HoverEvent.Value().(*c.Text)
	if !ok || interactiveOfficialPlainText(hoverText) != "Test" {
		t.Fatalf("hover value = %v, want Text{Test}", tr.S.HoverEvent.Value())
	}
}

func TestOfficialShowItemHover(t *testing.T) {
	for _, src := range []string{
		"<hover:show_item:'minecraft:stone':5>test",
		"<hover:show_item:'minecraft:stone':'5'>test",
	} {
		root := deserialize(t, src)
		if len(root.Extra) != 1 {
			t.Fatalf("%s: want 1 extra, got %d", src, len(root.Extra))
		}
		txt, ok := root.Extra[0].(*c.Text)
		if !ok || txt.Content != "test" {
			t.Fatalf("%s: extra[0] = %+v, want Text{test}", src, root.Extra[0])
		}
		hv := txt.S.HoverEvent
		if hv == nil || hv.Action().Name() != "show_item" {
			t.Fatalf("%s: hover = %v, want show_item", src, hv)
		}
		item, ok := hv.Value().(*c.ShowItemHoverType)
		if !ok {
			t.Fatalf("%s: hover value = %T, want *c.ShowItemHoverType", src, hv.Value())
		}
		if item.Item == nil || item.Item.Namespace() != "minecraft" || item.Item.Value() != "stone" {
			t.Fatalf("%s: item = %v, want minecraft:stone", src, item.Item)
		}
		if item.Count != 5 {
			t.Fatalf("%s: count = %d, want 5", src, item.Count)
		}
	}
}

func TestOfficialShowEntityHover(t *testing.T) {
	id := uuid.New()
	nameString := "<gold>Custom Name!"
	for _, src := range []string{
		"<hover:show_entity:'minecraft:zombie':" + id.String() + ":'" + nameString + "'>test",
		"<hover:show_entity:zombie:'" + id.String() + "':'" + nameString + "'>test",
	} {
		root := deserialize(t, src)
		if len(root.Extra) != 1 {
			t.Fatalf("%s: want 1 extra, got %d", src, len(root.Extra))
		}
		txt, ok := root.Extra[0].(*c.Text)
		if !ok || txt.Content != "test" {
			t.Fatalf("%s: extra[0] = %+v, want Text{test}", src, root.Extra[0])
		}
		hv := txt.S.HoverEvent
		if hv == nil || hv.Action().Name() != "show_entity" {
			t.Fatalf("%s: hover = %v, want show_entity", src, hv)
		}
		entity, ok := hv.Value().(*c.ShowEntityHoverType)
		if !ok {
			t.Fatalf("%s: hover value = %T, want *c.ShowEntityHoverType", src, hv.Value())
		}
		if entity.Type == nil || entity.Type.Namespace() != "minecraft" || entity.Type.Value() != "zombie" {
			t.Fatalf("%s: type = %v, want minecraft:zombie", src, entity.Type)
		}
		if entity.Id != id {
			t.Fatalf("%s: id = %v, want %v", src, entity.Id, id)
		}
		if entity.Name == nil {
			t.Fatalf("%s: name = nil, want gold Custom Name!", src)
		}
		nameLeaves := flattenComponent(entity.Name)
		wantLeaves := text("Custom Name!", mustColor(t, "gold").Hex())
		if len(nameLeaves) != len(wantLeaves) {
			t.Fatalf("%s: name leaves = %s, want %s", src, fmtLeaves(nameLeaves), fmtLeaves(wantLeaves))
		}
		for i := range wantLeaves {
			if nameLeaves[i].ch != wantLeaves[i].ch || nameLeaves[i].color != wantLeaves[i].color {
				t.Fatalf("%s: name leaf %d = %+v, want %+v", src, i, nameLeaves[i], wantLeaves[i])
			}
		}
	}
}

func TestOfficialStringPlaceholderInHover(t *testing.T) {
	root := deserialize(t, "<hover:show_text:'Word: <word>'><gold>Hover to see the word!",
		WithTagResolver(Placeholder("word", Text("Adventure"))))
	leaves := flattenComponent(root)
	want := text("Hover to see the word!", mustColor(t, "gold").Hex())
	if len(leaves) != len(want) {
		t.Fatalf("got %d chars, want %d\n got: %s\nwant: %s", len(leaves), len(want), fmtLeaves(leaves), fmtLeaves(want))
	}
	for i := range want {
		if leaves[i].ch != want[i].ch || leaves[i].color != want[i].color {
			t.Fatalf("char %d mismatch\n got: %+v\nwant: %+v", i, leaves[i], want[i])
		}
	}
	var hv c.HoverEvent
	if len(root.Extra) > 0 {
		hv = root.Extra[0].Style().HoverEvent
	}
	if hv == nil {
		t.Fatalf("no hover event found on rendered root: %+v", root)
	}
	hoverComp, ok := hv.Value().(c.Component)
	if !ok {
		t.Fatalf("hover value = %T, not a component", hv.Value())
	}
	if got := interactiveOfficialPlainText(hoverComp); got != "Word: Adventure" {
		t.Fatalf("hover text = %q, want %q", got, "Word: Adventure")
	}
}

func TestOfficialPhil(t *testing.T) {
	got := renderSrc(t, "<red><hover:show_text:'Message 1\nMessage 2'>My Message")
	want := text("My Message", mustColor(t, "red").Hex())
	if len(got) != len(want) {
		t.Fatalf("got %d chars, want %d\n got: %s\nwant: %s", len(got), len(want), fmtLeaves(got), fmtLeaves(want))
	}
	for i := range want {
		if got[i].ch != want[i].ch || got[i].color != want[i].color {
			t.Fatalf("char %d mismatch\n got: %+v\nwant: %+v", i, got[i], want[i])
		}
		if got[i].hover != "show_text:Message 1\nMessage 2" {
			t.Fatalf("char %d: hover = %q, want show_text:Message 1\\nMessage 2", i, got[i].hover)
		}
	}
}
