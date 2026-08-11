package micromessage

import (
	"strings"
	"testing"

	mccolor "go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
)

func checkLeaves(t *testing.T, got, want []charLeaf) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d chars, want %d\n got: %s\nwant: %s", len(got), len(want), fmtLeaves(got), fmtLeaves(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("char %d mismatch\n got: %+v\nwant: %+v\nfull got:  %s\nfull want: %s",
				i, got[i], want[i], fmtLeaves(got), fmtLeaves(want))
		}
	}
}

func deserialize(t *testing.T, src string, opts ...Option) *c.Text {
	t.Helper()
	root, err := Deserialize(src, opts...)
	if err != nil {
		t.Fatalf("deserialize error for %q: %v", src, err)
	}
	return root
}

func TestPlaceholderUnparsed(t *testing.T) {
	root := deserialize(t, "Hello <name/>!", WithTagResolver(Placeholder("name", Text("<b>Tom</b>"))))
	checkLeaves(t, flattenComponent(root), cat(text("Hello ", ""), text("<b>Tom</b>", ""), text("!", "")))
}

func TestPlaceholderParsed(t *testing.T) {
	root := deserialize(t, "Hello <name/>!", WithTagResolver(Placeholder("name", Parsed("<red>Tom</red>"))))
	checkLeaves(t, flattenComponent(root), cat(text("Hello ", ""), text("Tom", "#ff5555"), text("!", "")))
}

func TestPlaceholderComponent(t *testing.T) {
	comp := &c.Text{Content: "Tom", S: c.Style{Color: mustColor(t, "#00ff00")}}
	root := deserialize(t, "Hi <name/>!", WithTagResolver(Placeholder("name", ComponentTag(comp))))
	checkLeaves(t, flattenComponent(root), cat(text("Hi ", ""), text("Tom", "#00ff00"), text("!", "")))
}

func TestPlaceholderInheritsAmbientStyleWhenParsed(t *testing.T) {
	root := deserialize(t, "<bold>Hi <name/>!</bold>", WithTagResolver(Placeholder("name", Parsed("Tom"))))
	for _, l := range flattenComponent(root) {
		if !l.bold {
			t.Fatalf("expected bold to propagate into placeholder text, got %+v", flattenComponent(root))
		}
	}
}

func TestTagResolverFuncWithArgs(t *testing.T) {
	resolver := TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		if name != "score" {
			return Tag{}, false
		}
		a, ok := args.Pop()
		if !ok {
			return Tag{}, false
		}
		return Text("[" + a + "]"), true
	})
	root := deserialize(t, "<score:42/>", WithTagResolver(resolver))
	checkLeaves(t, flattenComponent(root), text("[42]", ""))
}

func TestArgumentQueuePopOr(t *testing.T) {
	resolver := TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		if name != "score" {
			return Tag{}, false
		}
		return Text("[" + args.PopOr("score needs an argument") + "]"), true
	})
	if _, err := Deserialize("<score/>", WithTagResolver(resolver)); err == nil {
		t.Fatal("expected an error from PopOr on a missing argument")
	} else if !strings.Contains(err.Error(), "score needs an argument") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnresolvedTagRendersAsLiteralText(t *testing.T) {
	// Matches the doc's own example: only StandardTags.Color() is enabled,
	// so <bold> is not recognized and renders as literal text.
	root := deserialize(t, "<green><bold>Hai", WithTags(StandardTags.Color()))
	got := flattenComponent(root)
	want := text("<bold>Hai", "#55ff55")
	checkLeaves(t, got, want)
}

func TestUnresolvedTagWithCloseRendersLiteralOpenAndClose(t *testing.T) {
	root := deserialize(t, "<unknown>hi</unknown>", WithTags(StandardTags.All()))
	got := flattenComponent(root)
	want := text("<unknown>hi</unknown>", "")
	checkLeaves(t, got, want)
}

func TestSelfReferentialPlaceholderDoesNotHang(t *testing.T) {
	if _, err := Deserialize("<loop/>", WithTagResolver(Placeholder("loop", Parsed("<loop/>")))); err != nil {
		t.Fatal(err)
	}
}

func TestPlaceholderInGradient(t *testing.T) {
	root := deserialize(t, "<gradient:red:blue>a<name/>b</gradient>", WithTagResolver(Placeholder("name", Text("X"))))
	got := flattenComponent(root)
	if len(got) != 3 || got[0].ch != 'a' || got[1].ch != 'X' || got[2].ch != 'b' {
		t.Fatalf("got %s", fmtLeaves(got))
	}
}

func TestPreprocessor(t *testing.T) {
	legacy := func(input string) string { return strings.ReplaceAll(input, "&c", "<red>") }
	root := deserialize(t, "&cHello", WithPreprocessor(legacy))
	checkLeaves(t, flattenComponent(root), text("Hello", "#ff5555"))
}

func TestMultiplePreprocessorsChain(t *testing.T) {
	first := func(input string) string { return strings.ReplaceAll(input, "&c", "<red>") }
	second := func(input string) string { return strings.ReplaceAll(input, "&l", "<bold>") }
	root := deserialize(t, "&c&lHi", WithPreprocessor(first), WithPreprocessor(second))
	got := flattenComponent(root)
	if len(got) != 2 || !got[0].bold || got[0].color != "#ff5555" {
		t.Fatalf("got %s", fmtLeaves(got))
	}
}

func TestBuilderColorOnlyDisablesBold(t *testing.T) {
	// Direct port of the doc's TagResolver.builder() example.
	root := deserialize(t, "<green><bold>Hai",
		WithTags(NewTagResolverBuilder().Resolver(StandardTags.Color()).Build()))
	checkLeaves(t, flattenComponent(root), text("<bold>Hai", "#55ff55"))
}

func TestParserDirectiveClearActsLikeReset(t *testing.T) {
	root := deserialize(t, "<red>hello <bold>world<clear>, how are you?",
		WithTags(StandardTags.All()), WithTagResolver(Resolver("clear", Reset)))
	got := flattenComponent(root)
	for _, l := range got[:len("hello world")] {
		_ = l
	}
	tail := got[len(got)-len(", how are you?"):]
	for _, l := range tail {
		if l.bold || l.color != "" {
			t.Fatalf("expected plain text after <clear>, got %+v", tail)
		}
	}
}

func TestCustomStylingTag(t *testing.T) {
	// Port of the doc's <a:link> example, using StylingTag.
	linkTag := TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		if name != "a" {
			return Tag{}, false
		}
		first := args.PopOr("the <a> tag requires exactly one argument, the link to open")
		link := strings.Join(append([]string{first}, args.Rest()...), ":")
		return StylingTag(
			ColorStyle(mustColor(t, "#5555ff")),
			DecorationStyle(c.Underlined, true),
			ClickStyle(c.NewClickEvent(c.ClickActions["open_url"], link)),
		), true
	})
	root := deserialize(t, "Hello, <a:https://example.com>click me!</a> but not me!", WithTagResolver(linkTag))
	got := flattenComponent(root)
	for _, l := range got[len("Hello, "):][:len("click me!")] {
		if !l.underlined || l.color != "#5555ff" || l.click != "open_url:https://example.com" {
			t.Fatalf("mismatch: %+v", l)
		}
	}
	for _, l := range got[len("Hello, click me!"):] {
		if l.underlined || l.color != "" {
			t.Fatalf("style leaked past </a>: %+v", l)
		}
	}
}

type upperCaseModifier struct{ chars int }

func (m *upperCaseModifier) Visit(n *Node) {
	if n.Kind == KindText {
		m.chars += len(n.Text)
	}
}
func (m *upperCaseModifier) PostVisit() {}
func (m *upperCaseModifier) Apply(current c.Component, _ int) c.Component {
	if txt, ok := current.(*c.Text); ok {
		txt.Content = strings.ToUpper(txt.Content)
	}
	return current
}

func TestCustomModifyingTag(t *testing.T) {
	upper := TagResolverFunc(func(name string, _ *ArgumentQueue) (Tag, bool) {
		if name != "upper" {
			return Tag{}, false
		}
		return ModifyingTagValue(&upperCaseModifier{}), true
	})
	root := deserialize(t, "<upper>hello</upper> world", WithTagResolver(upper))
	got := flattenComponent(root)
	word := string(runesOf(got[:5]))
	if word != "HELLO" {
		t.Fatalf("got %q, want HELLO", word)
	}
}

func runesOf(ls []charLeaf) []rune {
	rs := make([]rune, len(ls))
	for i, l := range ls {
		rs[i] = l.ch
	}
	return rs
}

func TestStrictModeErrorsOnUnclosedTag(t *testing.T) {
	if _, err := Deserialize("<bold>hi", WithStrict(true)); err == nil {
		t.Fatal("expected an error for an unclosed tag in strict mode")
	}
	if _, err := Deserialize("<bold>hi</bold>", WithStrict(true)); err != nil {
		t.Fatalf("closed tag should not error in strict mode: %v", err)
	}
	// Unresolved tags are still lenient even in strict mode.
	if _, err := Deserialize("<unknown/>", WithStrict(true)); err != nil {
		t.Fatalf("unresolved self-closing tag should not error: %v", err)
	}
}

func TestPresetNonInteractableStripsClick(t *testing.T) {
	root := deserialize(t, `<click:open_url:https://example.com>hi</click>`, NonInteractablePreset.Apply())
	got := flattenComponent(root)
	// <click> isn't in the non-interactable tag set, so it's literal text,
	// and the resulting component tree has no click events regardless.
	if strings.Contains(fmtLeaves(got), "click=") {
		t.Fatalf("expected no click event, got %s", fmtLeaves(got))
	}
}

func TestPresetFormattedTextDropsTranslation(t *testing.T) {
	root := deserialize(t, "<red>hi <lang:key/></red>",
		WithTags(StandardTags.All()), FormattedTextPreset.Apply())
	for _, ch := range root.Extra {
		if _, ok := ch.(*c.Translation); ok {
			t.Fatalf("expected translation component to be dropped, got %+v", root.Extra)
		}
	}
}

func mustColor(t *testing.T, hex string) mccolor.Color {
	t.Helper()
	col, ok := resolveColor(hex)
	if !ok {
		t.Fatalf("bad color %q", hex)
	}
	return col
}
