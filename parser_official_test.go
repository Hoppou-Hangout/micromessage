package micromessage

import (
	"testing"

	c "go.minekube.com/common/minecraft/component"
)

// Ported from adventure's MiniMessageParserTest.java / MiniMessageTest.java.
// Cases that hit confirmed real parser mismatches (see PR/summary notes) are
// intentionally omitted rather than ported with weakened assertions.

// officialLeaves builds charLeaf runs sharing one color/bold/click/hover,
// for the click+hover combos the shared text/textB helpers can't express.
func officialLeaves(s, color string, bold bool, click, hover string) []charLeaf {
	l := text(s, color)
	for i := range l {
		l[i].bold = bold
		l[i].click = click
		l[i].hover = hover
	}
	return l
}

// testInvalidTag
func TestOfficial_InvalidTag(t *testing.T) {
	check(t, "invalid_tag", `<red><test>`, text("<test>", "#ff5555"))
}

// testBackSpace
func TestOfficial_BackSpace(t *testing.T) {
	check(t, "backspace", `\!/ IMPORTANT \!/`, text(`\!/ IMPORTANT \!/`, ""))
}

// testGH5Quoted
func TestOfficial_GH5Quoted(t *testing.T) {
	src := `<dark_gray>»<gray> To download it from the internet, <click:open_url:"https://www.google.com"><hover:show_text:"<green>/!\ install it from Options/ResourcePacks in your game"><green><bold>CLICK HERE</bold></hover></click>`
	check(t, "gh5_quoted", src, cat(
		text("»", "#555555"),
		text(" To download it from the internet, ", "#aaaaaa"),
		officialLeaves("CLICK HERE", "#55ff55", true,
			"open_url:https://www.google.com",
			`show_text:/!\ install it from Options/ResourcePacks in your game`),
	))
}

// testDoubleNewLine
func TestOfficial_DoubleNewLine(t *testing.T) {
	check(t, "double_newline", "<red>Hello\n\nWorld", text("Hello\n\nWorld", "#ff5555"))
}

// testQuoteEscapingInArguments
func TestOfficial_QuoteEscapingInArguments(t *testing.T) {
	check(t, "double_quotes_single_quoted", `<lang:test:'""'>`, text(`""`, ""))
	check(t, "escaped_single_quotes_single_quoted", `<lang:test:'\'\''>`, text(`''`, ""))
	check(t, "single_quotes_double_quoted", `<lang:test:"''">`, text(`''`, ""))
	check(t, "escaped_double_quotes_double_quoted", `<lang:test:"\"\"">`, text(`""`, ""))
}

// testNoSwallowSpace (GH-111)
func TestOfficial_NoSwallowSpace(t *testing.T) {
	root := deserialize(t, `<red><hover:show_text:"Test"> <lang:item.minecraft.stick>`)
	got := flattenComponent(root)
	checkLeaves(t, got, officialLeaves(" ", "#ff5555", false, "", "show_text:Test"))

	found := false
	for _, ch := range root.Extra {
		if tr, ok := ch.(*c.Translation); ok && tr.Key == "item.minecraft.stick" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a translatable component for item.minecraft.stick")
	}
}

// testEscapeOutsideOfContext (GH-134)
func TestOfficial_EscapeOutsideOfContext(t *testing.T) {
	check(t, "escape_outside_of_context", `\"`, text(`\"`, ""))
}

// testEscapeInsideOfContext
func TestOfficial_EscapeInsideOfContext(t *testing.T) {
	root := deserialize(t, `<hover:show_text:'Look at\\ this \''>Test`)
	checkLeaves(t, flattenComponent(root), officialLeaves("Test", "", false, "", `show_text:Look at\ this '`))
}

// testEscapeAtEnd
func TestOfficial_EscapeAtEnd(t *testing.T) {
	check(t, "escape_at_end", `Please don't crash \`, text(`Please don't crash \`, ""))
}

// testEmptyTagPart (GH-166), quoted-empty-argument form
func TestOfficial_EmptyTagPart(t *testing.T) {
	check(t, "empty_tag_part_quoted", `<hover:show_text:"">text</hover>`,
		officialLeaves("text", "", false, "", "show_text:"))
}

// testValidTagNames
func TestOfficial_ValidTagNames(t *testing.T) {
	inputs := []string{
		"Hello <this_is_allowed> but cool?",
		"Hello <this-is-allowed> but cool?",
		"Hello <!allowed> but cool?",
		"Hello <?allowed> but cool?",
		"Hello <#allowed> but cool?",
	}
	for _, in := range inputs {
		if _, err := Deserialize(in); err != nil {
			t.Errorf("Deserialize(%q) unexpected error: %v", in, err)
		}
	}
}

// testNegatedDecorationTags (GH-1408)
func TestOfficial_NegatedDecorationTags(t *testing.T) {
	src := `<!i>Not italic<!b>Not bold also</!b>back to italic.`
	got := renderSrc(t, src)
	want := cat(
		mkLeaves("Not italic", leafMods{italic: boolPtrLit(false)}),
		mkLeaves("Not bold also", leafMods{italic: boolPtrLit(false), bold: boolPtrLit(false)}),
		mkLeaves("back to italic.", leafMods{italic: boolPtrLit(false)}),
	)
	compareLeavesWithBoolPtrs(t, src, got, want)
}

// allClosedTagsStrict
func TestOfficial_StrictAllClosedTagsOK(t *testing.T) {
	src := `<red>RED<green>GREEN</green>RED<blue>BLUE</blue></red>`
	root := deserialize(t, src, WithStrict(true))
	checkLeaves(t, flattenComponent(root), cat(
		text("RED", "#ff5555"),
		text("GREEN", "#55ff55"),
		text("RED", "#ff5555"),
		text("BLUE", "#5555ff"),
	))
}

// unclosedTagStrict (message text not asserted, only that it errors)
func TestOfficial_StrictUnclosedTagErrors(t *testing.T) {
	src := `<red>RED<green>GREEN</green>RED<blue>BLUE`
	if _, err := Deserialize(src, WithStrict(true)); err == nil {
		t.Fatalf("expected an error for unclosed tag in strict mode, got none")
	}
}

// testNonStrict
func TestOfficial_NonStrictClickInsideColor(t *testing.T) {
	src := `<gray>Example: <click:suggest_command:/plot flag set coral-dry true><gold>/plot flag set coral-dry true</gold></click></gray>`
	root := deserialize(t, src)
	want := cat(
		text("Example: ", "#aaaaaa"),
		officialLeaves("/plot flag set coral-dry true", "#ffaa00", false, "suggest_command:/plot flag set coral-dry true", ""),
	)
	checkLeaves(t, flattenComponent(root), want)
}

// testPreprocessing
func TestOfficial_Preprocessing(t *testing.T) {
	root := deserialize(t, "Hello", WithPreprocessor(func(s string) string {
		return "<red>" + s + ", world!</red>"
	}))
	checkLeaves(t, flattenComponent(root), text("Hello, world!", "#ff5555"))
}

// testCustomRegistry
func TestOfficial_CustomRegistryEmpty(t *testing.T) {
	root := deserialize(t, `<green><bold><test>`,
		WithTags(NewTagResolverBuilder().Build()),
		WithTagResolver(Placeholder("test", Text("TEST"))))
	checkLeaves(t, flattenComponent(root), cat(
		text("<green><bold>", ""),
		text("TEST", ""),
	))
}

// testCustomRegistryBuilder
func TestOfficial_CustomRegistryColorOnly(t *testing.T) {
	root := deserialize(t, `<green><bold><test>`,
		WithTags(StandardTags.Color()),
		WithTagResolver(Placeholder("test", Text("TEST"))))
	checkLeaves(t, flattenComponent(root), text("<bold>TEST", "#55ff55"))
}

// testNodesInPlaceholder
func TestOfficial_NodesInPlaceholderNotReparsed(t *testing.T) {
	root := deserialize(t, `<red><username><gray>: <red><message>`,
		WithTagResolver(Placeholder("username", Text("MiniDigger"))),
		WithTagResolver(Placeholder("message", Text("</pre><red>Test"))))
	want := cat(
		text("MiniDigger", "#ff5555"),
		text(": ", "#aaaaaa"),
		text("</pre><red>Test", "#ff5555"),
	)
	checkLeaves(t, flattenComponent(root), want)
}
