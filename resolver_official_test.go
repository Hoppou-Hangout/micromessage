package micromessage

// Ported from Adventure's TagResolverTest.java, PlaceholderTest.java,
// PreProcessTagTest.java, and MiniMessagePresetTest.java.

import (
	"testing"

	c "go.minekube.com/common/minecraft/component"
)

// -- TagResolverTest ---------------------------------------------------

// testContextParseOne / testContextParseVarargs
func TestOfficialResolverSingleArgument(t *testing.T) {
	root := deserialize(t, "<foo> <bar><green>!</green>", WithTagResolver(Placeholder("foo", Parsed("<red>Hello</red>").SelfClosing())))
	checkLeaves(t, flattenComponent(root),
		cat(text("Hello", "#ff5555"), text(" <bar>", ""), text("!", "#55ff55")))
}

func TestOfficialResolverVarargsArguments(t *testing.T) {
	root := deserialize(t, "<foo> <bar><green>!</green>",
		WithTagResolver(Placeholder("foo", Parsed("<red>Hello</red>").SelfClosing())),
		WithTagResolver(Placeholder("bar", Parsed("<yellow>World</yellow>").SelfClosing())))
	checkLeaves(t, flattenComponent(root),
		cat(text("Hello", "#ff5555"), text(" ", ""), text("World", "#ffff55"), text("!", "#55ff55")))
}

// testSingleAndResolversCombine, adapted: TagResolverBuilder priority --
// real Adventure's TagResolver.builder().build() gives priority to the
// LAST-added resolver for an overlapping name (TagResolverBuilderImpl
// reverses the list before wrapping it in a first-match-wins
// SequentialTagResolver). Our TagResolverBuilder.Build() does not reverse:
// it keeps first-added-wins order (see its doc comment in resolver.go).
// This test documents that mismatch: it currently FAILS against our
// implementation.
func TestOfficialResolverBuilderPriorityLastAddedWins(t *testing.T) {
	placeholders := NewTagResolverBuilder().
		Resolver(Placeholder("foo", Text("fizz"))).
		Resolver(Placeholder("overlapping", Text("from list"))).
		Build()
	fromFunc := TagResolverFunc(func(name string, _ *ArgumentQueue) (Tag, bool) {
		switch name {
		case "one":
			return Text("fish"), true
		case "overlapping":
			return Text("from resolver"), true
		}
		return Tag{}, false
	})
	built := NewTagResolverBuilder().Resolver(placeholders).Resolver(fromFunc).Build()

	root := deserialize(t, "<foo> <one> <overlapping>", WithTagResolver(built))
	got := flattenComponent(root)
	want := cat(text("fizz", ""), text(" ", ""), text("fish", ""), text(" ", ""), text("from resolver", ""))
	checkLeaves(t, got, want)
}

// -- PlaceholderTest -----------------------------------------------------

func TestOfficialCheckPlaceholder(t *testing.T) {
	root := deserialize(t, "<test>", WithTagResolver(Placeholder("test", Parsed("Hello!").SelfClosing())))
	checkLeaves(t, flattenComponent(root), text("Hello!", ""))
}

// testPlaceholderOrder: nested ambient color/decoration should apply to
// Text-kind placeholders that inherit ambient style (the closest analogue
// to Java's Placeholder.component, whose inserted component visually
// inherits ambient color through Adventure's component-tree style
// resolution; our ComponentTag deliberately does NOT do this per its doc
// comment, so Text is used here to exercise the ordering/inheritance
// behavior instead).
func TestOfficialPlaceholderOrder(t *testing.T) {
	root := deserialize(t, "<gray><arg1><red><arg2> <arg3> <arg4>",
		WithTagResolver(NewTagResolverBuilder().
			Resolver(Placeholder("arg1", Text("ONE"))).
			Resolver(Placeholder("arg2", Text("TWO"))).
			Resolver(Placeholder("arg3", Text("THREE"))).
			Resolver(Placeholder("arg4", Text("FOUR"))).
			Build()))
	got := flattenComponent(root)
	want := cat(
		text("ONE", "#aaaaaa"),
		text("TWO", "#ff5555"), text(" ", "#ff5555"), text("THREE", "#ff5555"), text(" ", "#ff5555"), text("FOUR", "#ff5555"),
	)
	checkLeaves(t, got, want)
}

func TestOfficialPlaceholderOrder2(t *testing.T) {
	root := deserialize(t, "<gray><arg1></gray><red><arg2></red><blue><arg3></blue> <green><arg4>",
		WithTagResolver(NewTagResolverBuilder().
			Resolver(Placeholder("arg1", Text("ONE"))).
			Resolver(Placeholder("arg2", Text("TWO"))).
			Resolver(Placeholder("arg3", Text("THREE"))).
			Resolver(Placeholder("arg4", Text("FOUR"))).
			Build()))
	got := flattenComponent(root)
	want := cat(
		text("ONE", "#aaaaaa"),
		text("TWO", "#ff5555"),
		text("THREE", "#5555ff"),
		text(" ", ""),
		text("FOUR", "#55ff55"),
	)
	checkLeaves(t, got, want)
}

// testRepeatedResolvingOfStringPlaceholders: nested placeholders resolve
// recursively (feline resolves inside animal's Parsed value). The Java
// original uses an unclosed "<red><feline>" substitution whose color bleeds
// across the placeholder boundary into the surrounding " makes a sound"
// text -- that relies on Java's source-level splice-then-relex, which
// Parsed's doc comment explicitly calls out as a case where our
// independent-subdocument approach can differ (it renders the substitution
// as its own document, so an unclosed tag inside it does not bleed past the
// substitution). This uses a self-contained closed substitution instead, to
// exercise the recursive-resolution behavior without depending on that
// documented divergence.
func TestOfficialRepeatedResolvingOfPlaceholders(t *testing.T) {
	root := deserialize(t, "<animal> makes a sound", WithTagResolver(NewTagResolverBuilder().
		Resolver(Placeholder("animal", Parsed("<red><feline></red>").SelfClosing())).
		Resolver(Placeholder("feline", Text("cat"))).
		Build()))
	got := flattenComponent(root)
	want := cat(text("cat", "#ff5555"), text(" makes a sound", ""))
	checkLeaves(t, got, want)
}

// A custom TagResolver registered via WithTagResolver is the closest
// analogue to a resolver passed directly to Adventure's
// MiniMessage#deserialize(input, tagResolver): in real Adventure that
// per-call resolver takes priority over the parser's base/standard tags
// (MiniMessageParser combines them as TagResolver.resolver(base, extra),
// and TagResolver.builder() gives priority to the later-added argument).
// Our resolvers() (resolver.go) instead always tries the base/standard tag
// set before WithTagResolver extras, so a placeholder can never shadow a
// built-in tag name. This test documents that mismatch: it currently FAILS
// against our implementation.
func TestOfficialExtraResolverOverridesBuiltin(t *testing.T) {
	root := deserialize(t, "<red>foo", WithTagResolver(Placeholder("red", Text("X"))))
	checkLeaves(t, flattenComponent(root), cat(text("X", ""), text("foo", "")))
}

// -- PreProcessTagTest -----------------------------------------------------

func TestOfficialCheckPreProcessTag(t *testing.T) {
	resolver := TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		if name != "test" {
			return Tag{}, false
		}
		a, _ := args.Pop()
		b, _ := args.Pop()
		return Parsed(a + b).SelfClosing(), true
	})
	root := deserialize(t, "<test:'Hello! ':bla>", WithTagResolver(resolver))
	checkLeaves(t, flattenComponent(root), text("Hello! bla", ""))
}

func TestOfficialCheckSpecialChars(t *testing.T) {
	resolver := TagResolverFunc(func(name string, args *ArgumentQueue) (Tag, bool) {
		if name != "test" {
			return Tag{}, false
		}
		a, _ := args.Pop()
		b, _ := args.Pop()
		return Parsed(a + b).SelfClosing(), true
	})
	root := deserialize(t, "<test:'::':'<bla>'>", WithTagResolver(resolver))
	checkLeaves(t, flattenComponent(root), text("::<bla>", ""))
}

// -- MiniMessagePresetTest -----------------------------------------------

func TestOfficialPresetDefault(t *testing.T) {
	root := deserialize(t,
		"<red>Hello <blue><bold>world</bold></blue></red><green><click:run_command:/test>!</click></green>",
		DefaultPreset.Apply())
	got := flattenComponent(root)
	want := cat(
		text("Hello ", "#ff5555"),
		textB("world", "#5555ff"),
	)
	want = append(want, charLeaf{ch: '!', color: "#55ff55", click: "run_command:/test"})
	checkLeaves(t, got, want)
}

func TestOfficialPresetNonInteractableLiteralTags(t *testing.T) {
	src := "<click:run_command:/test>No click</click><hover:show_text:hover>No hover</hover><insert:insert>No insertion</insert>"
	root := deserialize(t, src, NonInteractablePreset.Apply())
	got := flattenComponent(root)
	checkLeaves(t, got, text(src, ""))
}

func TestOfficialPresetNonInteractableFiltersCustomTag(t *testing.T) {
	dangerous := TagResolverFunc(func(name string, _ *ArgumentQueue) (Tag, bool) {
		if name != "dangerous" {
			return Tag{}, false
		}
		return StylingTag(ClickStyle(c.NewClickEvent(c.ClickActions["run_command"], "/kill"))), true
	})
	root := deserialize(t, "<dangerous>test</dangerous>", NonInteractablePreset.Apply(), WithTagResolver(dangerous))
	checkLeaves(t, flattenComponent(root), text("test", ""))
}

func TestOfficialPresetFormattedTextBasicFormatting(t *testing.T) {
	root := deserialize(t, "<red>Hello <blue><bold>world</bold></blue></red>", FormattedTextPreset.Apply())
	got := flattenComponent(root)
	want := cat(text("Hello ", "#ff5555"), textB("world", "#5555ff"))
	checkLeaves(t, got, want)
}

func TestOfficialPresetFormattedTextLiteralClick(t *testing.T) {
	src := "<click:run_command:/test>No click</click>"
	root := deserialize(t, src, FormattedTextPreset.Apply())
	checkLeaves(t, flattenComponent(root), text(src, ""))
}

func TestOfficialPresetFormattedTextLiteralKeybind(t *testing.T) {
	src := "<keybind:key.jump>"
	root := deserialize(t, src, FormattedTextPreset.Apply())
	checkLeaves(t, flattenComponent(root), text(src, ""))
}

func TestOfficialPresetFormattedTextFiltersNonTextCustomTag(t *testing.T) {
	mykey := TagResolverFunc(func(name string, _ *ArgumentQueue) (Tag, bool) {
		if name != "mykey" {
			return Tag{}, false
		}
		return ComponentTag(&c.Keybind{Key: "key.jump"}), true
	})
	root := deserialize(t, "<mykey>", FormattedTextPreset.Apply(), WithTagResolver(mykey))
	got := flattenComponent(root)
	if len(got) != 0 {
		t.Fatalf("expected keybind component to be dropped, got %s", fmtLeaves(got))
	}
}

func TestOfficialPresetFormattedTextFiltersInteractiveCustomText(t *testing.T) {
	dangertext := TagResolverFunc(func(name string, _ *ArgumentQueue) (Tag, bool) {
		if name != "dangertext" {
			return Tag{}, false
		}
		return ComponentTag(&c.Text{
			Content: "danger",
			S:       c.Style{ClickEvent: c.NewClickEvent(c.ClickActions["run_command"], "/kill")},
		}), true
	})
	root := deserialize(t, "<dangertext>", FormattedTextPreset.Apply(), WithTagResolver(dangertext))
	got := flattenComponent(root)
	checkLeaves(t, got, text("danger", ""))
	if got[0].click != "" {
		t.Fatalf("expected click event to be stripped, got %+v", got[0])
	}
}

func TestOfficialPresetFormattedTextFiltersNonTextChild(t *testing.T) {
	nontextchild := TagResolverFunc(func(name string, _ *ArgumentQueue) (Tag, bool) {
		if name != "nontextchild" {
			return Tag{}, false
		}
		return ComponentTag(&c.Text{
			Content: "parent",
			Extra:   []c.Component{&c.Keybind{Key: "key.jump"}},
		}), true
	})
	root := deserialize(t, "<nontextchild>", FormattedTextPreset.Apply(), WithTagResolver(nontextchild))
	checkLeaves(t, flattenComponent(root), text("parent", ""))
}
