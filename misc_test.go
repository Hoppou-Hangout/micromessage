package micromessage

import "testing"

// --- Ported from ColorTagTest.java -----------------------------------

func TestColor_Basic(t *testing.T) {
	check(t, "yellow_green_nested", `<yellow>TEST<green> nested</green>Test`,
		cat(text("TEST", "#ffff55"), text(" nested", "#55ff55"), text("Test", "#ffff55")))

	check(t, "yellow_green_reopen_yellow", `<yellow>TEST<green> nested<yellow>Test`,
		cat(text("TEST", "#ffff55"), text(" nested", "#55ff55"), text("Test", "#ffff55")))
}

func TestColor_BritishAliases(t *testing.T) {
	grey := renderSrc(t, `<grey>This is english`)
	gray := renderSrc(t, `<gray>This is english`)
	if fmtLeaves(grey) != fmtLeaves(gray) {
		t.Fatalf("grey != gray: %s vs %s", fmtLeaves(grey), fmtLeaves(gray))
	}
	darkGrey := renderSrc(t, `<dark_grey>This is still english`)
	darkGray := renderSrc(t, `<dark_gray>This is still english`)
	if fmtLeaves(darkGrey) != fmtLeaves(darkGray) {
		t.Fatalf("dark_grey != dark_gray: %s vs %s", fmtLeaves(darkGrey), fmtLeaves(darkGray))
	}
}

func TestColor_ExplicitTagForm(t *testing.T) {
	check(t, "color_yellow_green", `<color:yellow>TEST<color:green> nested</color:green>Test`,
		cat(text("TEST", "#ffff55"), text(" nested", "#55ff55"), text("Test", "#ffff55")))
}

func TestColor_HexForms(t *testing.T) {
	check(t, "color_hex", `<color:#ff00ff>TEST<color:#00ff00> nested</color:#00ff00>Test`,
		cat(text("TEST", "#ff00ff"), text(" nested", "#00ff00"), text("Test", "#ff00ff")))

	check(t, "bare_hex", `<#ff00ff>TEST<#00ff00> nested</#00ff00>Test`,
		cat(text("TEST", "#ff00ff"), text(" nested", "#00ff00"), text("Test", "#ff00ff")))

	check(t, "c_alias_hex", `<c:#ff00ff>TEST<c:#00ff00> nested</c>Test`,
		cat(text("TEST", "#ff00ff"), text(" nested", "#00ff00"), text("Test", "#ff00ff")))
}

func TestColor_AllAliasesAgree(t *testing.T) {
	a := renderSrc(t, `<color:red>x</color>`)
	b := renderSrc(t, `<colour:red>x</colour>`)
	cc := renderSrc(t, `<c:red>x</c>`)
	if fmtLeaves(a) != fmtLeaves(b) || fmtLeaves(b) != fmtLeaves(cc) {
		t.Fatalf("aliases disagree: %s / %s / %s", fmtLeaves(a), fmtLeaves(b), fmtLeaves(cc))
	}
}

// --- Ported from DecorationTagTest.java --------------------------------

func TestDecoration_ExplicitFalseArg(t *testing.T) {
	// <italic:false>Test<bold:false>Test2<bold>Test3
	got := renderSrc(t, `<italic:false>Test<bold:false>Test2<bold>Test3`)
	want := []charLeaf{}
	want = append(want, mkLeaves("Test", leafMods{italic: boolPtrLit(false)})...)
	want = append(want, mkLeaves("Test2", leafMods{italic: boolPtrLit(false), bold: boolPtrLit(false)})...)
	want = append(want, mkLeaves("Test3", leafMods{italic: boolPtrLit(false), bold: boolPtrLit(true)})...)
	compareLeavesWithBoolPtrs(t, `<italic:false>Test<bold:false>Test2<bold>Test3`, got, want)
}

func TestDecoration_NegationShorthand(t *testing.T) {
	got := renderSrc(t, `<!italic>Test<!bold>Test2<bold>Test3`)
	want := []charLeaf{}
	want = append(want, mkLeaves("Test", leafMods{italic: boolPtrLit(false)})...)
	want = append(want, mkLeaves("Test2", leafMods{italic: boolPtrLit(false), bold: boolPtrLit(false)})...)
	want = append(want, mkLeaves("Test3", leafMods{italic: boolPtrLit(false), bold: boolPtrLit(true)})...)
	compareLeavesWithBoolPtrs(t, `<!italic>Test<!bold>Test2<bold>Test3`, got, want)
}

// leafMods/mkLeaves/compareLeavesWithBoolPtrs let us test decoration *false*
// explicitly (charLeaf's plain bool fields can't distinguish "not mentioned"
// from "explicitly false", but Adventure's own semantics only care about final
// true/false here, so we just compare bools).
type leafMods struct {
	bold, italic *bool
}

func boolPtrLit(b bool) *bool { return &b }

func mkLeaves(s string, m leafMods) []charLeaf {
	l := text(s, "")
	for i := range l {
		if m.bold != nil {
			l[i].bold = *m.bold
		}
		if m.italic != nil {
			l[i].italic = *m.italic
		}
	}
	return l
}

func compareLeavesWithBoolPtrs(t *testing.T, src string, got, want []charLeaf) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %d chars, want %d\n got: %s\nwant: %s", src, len(got), len(want), fmtLeaves(got), fmtLeaves(want))
	}
	for i := range want {
		if got[i].ch != want[i].ch || got[i].bold != want[i].bold || got[i].italic != want[i].italic {
			t.Fatalf("%s: char %d mismatch\n got: %+v\nwant: %+v\nfull got:  %s\nfull want: %s",
				src, i, got[i], want[i], fmtLeaves(got), fmtLeaves(want))
		}
	}
}

// --- Ported from ClickTagTest.java -------------------------------------

func TestClick_Basic(t *testing.T) {
	got := renderSrc(t, `<click:run_command:test>TEST`)
	want := []charLeaf{}
	for _, r := range "TEST" {
		want = append(want, charLeaf{ch: r, click: "run_command:test"})
	}
	if fmtLeaves(got) != fmtLeaves(want) {
		t.Fatalf("got %s want %s", fmtLeaves(got), fmtLeaves(want))
	}
}

func TestClick_UnquotedArgumentWithSpace(t *testing.T) {
	// testClickExtendedCommand: unquoted args may contain spaces up to the
	// closing '>' - this only works if tag-state whitespace isn't elided.
	got := renderSrc(t, `<click:run_command:/test command>TEST`)
	want := []charLeaf{}
	for _, r := range "TEST" {
		want = append(want, charLeaf{ch: r, click: "run_command:/test command"})
	}
	if fmtLeaves(got) != fmtLeaves(want) {
		t.Fatalf("got %s want %s", fmtLeaves(got), fmtLeaves(want))
	}
}

// --- Ported from HoverTagTest.java --------------------------------------

func TestHover_Basic(t *testing.T) {
	got := renderSrc(t, `<hover:show_text:"a plain hover">TEST`)
	want := []charLeaf{}
	for _, r := range "TEST" {
		want = append(want, charLeaf{ch: r, hover: "show_text:a plain hover"})
	}
	if fmtLeaves(got) != fmtLeaves(want) {
		t.Fatalf("got %s want %s", fmtLeaves(got), fmtLeaves(want))
	}
}

// --- Ported from MiniMessageParserTest.java (general behavior) --------

func TestGeneral_MismatchedTags(t *testing.T) {
	// testMismatchedTags: a close tag that doesn't match anything currently
	// open doesn't close anything -- it's literal text, and "green" (never
	// explicitly closed) auto-closes at EOF instead.
	check(t, "mismatched", `<green>hello</red>`, text("hello</red>", "#55ff55"))
}

func TestGeneral_CaseInsensitiveTagNames(t *testing.T) {
	check(t, "bold_uppercase", `<red>this is <BOLD>an error</bold> message`,
		cat(text("this is ", "#ff5555"), textB("an error", "#ff5555"), text(" message", "#ff5555")))

	check(t, "c_red_uppercase", `<C:reD>also red`, text("also red", "#ff5555"))
}

func TestGeneral_SelfClosingIgnorable(t *testing.T) {
	// testIgnorableSelfClosable: <red/>things - the self-close carries no text
	// of its own, so nothing should render with red at all; "things" is plain.
	check(t, "self_close_no_text", `<red/>things`, text("things", ""))
}
