package micromessage

import (
	"fmt"
	"testing"

	c "go.minekube.com/common/minecraft/component"
)

// --- Ported from ColorTagTest.java --------------------------------------

func TestOfficialColor_Basic(t *testing.T) {
	check(t, "yellow_green_close", `<yellow>TEST<green> nested</green>Test`,
		cat(text("TEST", "#ffff55"), text(" nested", "#55ff55"), text("Test", "#ffff55")))

	check(t, "yellow_green_reopen_yellow", `<yellow>TEST<green> nested<yellow>Test`,
		cat(text("TEST", "#ffff55"), text(" nested", "#55ff55"), text("Test", "#ffff55")))
}

func TestOfficialColor_British(t *testing.T) {
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

func TestOfficialColor_BritishColour(t *testing.T) {
	a := renderSrc(t, `<colour:grey>This is english`)
	b := renderSrc(t, `<color:gray>This is english`)
	if fmtLeaves(a) != fmtLeaves(b) {
		t.Fatalf("colour:grey != color:gray: %s vs %s", fmtLeaves(a), fmtLeaves(b))
	}
}

func TestOfficialColor_NewColorTagForm(t *testing.T) {
	check(t, "color_yellow_green_close", `<color:yellow>TEST<color:green> nested</color:green>Test`,
		cat(text("TEST", "#ffff55"), text(" nested", "#55ff55"), text("Test", "#ffff55")))

	check(t, "color_yellow_green_reopen", `<color:yellow>TEST<color:green> nested<color:yellow>Test`,
		cat(text("TEST", "#ffff55"), text(" nested", "#55ff55"), text("Test", "#ffff55")))
}

func TestOfficialColor_HexColor(t *testing.T) {
	check(t, "hex_close", `<color:#ff00ff>TEST<color:#00ff00> nested</color:#00ff00>Test`,
		cat(text("TEST", "#ff00ff"), text(" nested", "#00ff00"), text("Test", "#ff00ff")))

	check(t, "hex_reopen", `<color:#ff00ff>TEST<color:#00ff00> nested<color:#ff00ff>Test`,
		cat(text("TEST", "#ff00ff"), text(" nested", "#00ff00"), text("Test", "#ff00ff")))
}

func TestOfficialColor_HexColorShort(t *testing.T) {
	check(t, "bare_hex_close", `<#ff00ff>TEST<#00ff00> nested</#00ff00>Test`,
		cat(text("TEST", "#ff00ff"), text(" nested", "#00ff00"), text("Test", "#ff00ff")))

	check(t, "bare_hex_reopen", `<#ff00ff>TEST<#00ff00> nested<#ff00ff>Test`,
		cat(text("TEST", "#ff00ff"), text(" nested", "#00ff00"), text("Test", "#ff00ff")))
}

func TestOfficialColor_HexColorC(t *testing.T) {
	check(t, "c_hex_close", `<c:#ff00ff>TEST<c:#00ff00> nested</c>Test`,
		cat(text("TEST", "#ff00ff"), text(" nested", "#00ff00"), text("Test", "#ff00ff")))

	check(t, "c_hex_reopen", `<c:#ff00ff>TEST<c:#00ff00> nested<c:#ff00ff>Test`,
		cat(text("TEST", "#ff00ff"), text(" nested", "#00ff00"), text("Test", "#ff00ff")))
}

func TestOfficialColor_AllAliases(t *testing.T) {
	check(t, "color_hex", `<color:#ff00ff>AGGRESSIVE TEST</color>`, text("AGGRESSIVE TEST", "#ff00ff"))
	check(t, "colour_hex", `<colour:#00ffff>less aggressive test</colour>`, text("less aggressive test", "#00ffff"))
	check(t, "c_hex", `<c:#1234de>Mildly Aggressive Test</c>`, text("Mildly Aggressive Test", "#1234de"))
	check(t, "color_named", `<color:red>AGGRESSIVE TEST</color>`, text("AGGRESSIVE TEST", "#ff5555"))
	check(t, "colour_named", `<colour:green>less aggressive test</colour>`, text("less aggressive test", "#55ff55"))
	check(t, "c_named", `<c:blue>Mildly Aggressive Test</c>`, text("Mildly Aggressive Test", "#5555ff"))
}

func TestOfficialColor_Simple(t *testing.T) {
	check(t, "yellow_simple", `<yellow>TEST`, text("TEST", "#ffff55"))
}

// --- Ported from DecorationTagTest.java ---------------------------------

// styleOfficialLeafMods lets us assert decoration *false* explicitly, since
// charLeaf's plain bools can't distinguish "not mentioned" (defaults false)
// from "explicitly false" -- but for these tests only the final true/false
// state matters, so this is a thin wrapper.
type styleOfficialLeafMods struct {
	bold, italic, underlined *bool
}

func styleOfficialBoolPtr(b bool) *bool { return &b }

func styleOfficialMkLeaves(s string, m styleOfficialLeafMods) []charLeaf {
	l := text(s, "")
	for i := range l {
		if m.bold != nil {
			l[i].bold = *m.bold
		}
		if m.italic != nil {
			l[i].italic = *m.italic
		}
		if m.underlined != nil {
			l[i].underlined = *m.underlined
		}
	}
	return l
}

func styleOfficialCompareDecoration(t *testing.T, src string, got, want []charLeaf) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %d chars, want %d\n got: %s\nwant: %s", src, len(got), len(want), fmtLeaves(got), fmtLeaves(want))
	}
	for i := range want {
		if got[i].ch != want[i].ch || got[i].bold != want[i].bold || got[i].italic != want[i].italic || got[i].underlined != want[i].underlined {
			t.Fatalf("%s: char %d mismatch\n got: %+v\nwant: %+v\nfull got:  %s\nfull want: %s",
				src, i, got[i], want[i], fmtLeaves(got), fmtLeaves(want))
		}
	}
}

func TestOfficialDecoration_SerializeNested(t *testing.T) {
	// testSerializeDecoration: <underlined>This is <bold>underlined</bold></underlined>, this isn't
	got := renderSrc(t, `<underlined>This is <bold>underlined</bold></underlined>, this isn't`)
	want := []charLeaf{}
	want = append(want, styleOfficialMkLeaves("This is ", styleOfficialLeafMods{underlined: styleOfficialBoolPtr(true)})...)
	want = append(want, styleOfficialMkLeaves("underlined", styleOfficialLeafMods{underlined: styleOfficialBoolPtr(true), bold: styleOfficialBoolPtr(true)})...)
	want = append(want, styleOfficialMkLeaves(", this isn't", styleOfficialLeafMods{})...)
	styleOfficialCompareDecoration(t, `<underlined>This is <bold>underlined</bold></underlined>, this isn't`, got, want)
}

func TestOfficialDecoration_SerializeNegated(t *testing.T) {
	// testSerializeDecorationNegated: <!underlined>Not underlined<!bold>not bold<underlined>underlined</underlined></!bold> not underlined
	src := `<!underlined>Not underlined<!bold>not bold<underlined>underlined</underlined></!bold> not underlined`
	got := renderSrc(t, src)
	want := []charLeaf{}
	want = append(want, styleOfficialMkLeaves("Not underlined", styleOfficialLeafMods{underlined: styleOfficialBoolPtr(false)})...)
	want = append(want, styleOfficialMkLeaves("not bold", styleOfficialLeafMods{underlined: styleOfficialBoolPtr(false), bold: styleOfficialBoolPtr(false)})...)
	want = append(want, styleOfficialMkLeaves("underlined", styleOfficialLeafMods{underlined: styleOfficialBoolPtr(true), bold: styleOfficialBoolPtr(false)})...)
	want = append(want, styleOfficialMkLeaves(" not underlined", styleOfficialLeafMods{underlined: styleOfficialBoolPtr(false)})...)
	styleOfficialCompareDecoration(t, src, got, want)
}

func TestOfficialDecoration_DisabledExplicit(t *testing.T) {
	// testDisabledDecoration: <italic:false>Test<bold:false>Test2<bold>Test3
	src := `<italic:false>Test<bold:false>Test2<bold>Test3`
	got := renderSrc(t, src)
	want := []charLeaf{}
	want = append(want, styleOfficialMkLeaves("Test", styleOfficialLeafMods{italic: styleOfficialBoolPtr(false)})...)
	want = append(want, styleOfficialMkLeaves("Test2", styleOfficialLeafMods{italic: styleOfficialBoolPtr(false), bold: styleOfficialBoolPtr(false)})...)
	want = append(want, styleOfficialMkLeaves("Test3", styleOfficialLeafMods{italic: styleOfficialBoolPtr(false), bold: styleOfficialBoolPtr(true)})...)
	styleOfficialCompareDecoration(t, src, got, want)
}

func TestOfficialDecoration_DisabledShorthand(t *testing.T) {
	// testDisabledDecorationShorthand: <!italic>Test<!bold>Test2<bold>Test3
	src := `<!italic>Test<!bold>Test2<bold>Test3`
	got := renderSrc(t, src)
	want := []charLeaf{}
	want = append(want, styleOfficialMkLeaves("Test", styleOfficialLeafMods{italic: styleOfficialBoolPtr(false)})...)
	want = append(want, styleOfficialMkLeaves("Test2", styleOfficialLeafMods{italic: styleOfficialBoolPtr(false), bold: styleOfficialBoolPtr(false)})...)
	want = append(want, styleOfficialMkLeaves("Test3", styleOfficialLeafMods{italic: styleOfficialBoolPtr(false), bold: styleOfficialBoolPtr(true)})...)
	styleOfficialCompareDecoration(t, src, got, want)
}

func TestOfficialDecoration_ErrorOnShorthandAndLonghand(t *testing.T) {
	// testErrorOnShorthandAndLongHand: <!italic:true> mixes shorthand negation
	// with an explicit arg -- real MiniMessage treats the whole tag as
	// unresolved and renders it literally.
	check(t, "shorthand_and_longhand_literal", `<!italic:true>Go decide on something, god dammit!`,
		text(`<!italic:true>Go decide on something, god dammit!`, ""))
}

func TestOfficialDecoration_ShorthandClosing(t *testing.T) {
	// testDecorationShorthandClosing: <italic:false>Hello! <italic>spooky</italic> not spooky</italic:false>
	src := `<italic:false>Hello! <italic>spooky</italic> not spooky</italic:false>`
	got := renderSrc(t, src)
	want := []charLeaf{}
	want = append(want, styleOfficialMkLeaves("Hello! ", styleOfficialLeafMods{italic: styleOfficialBoolPtr(false)})...)
	want = append(want, styleOfficialMkLeaves("spooky", styleOfficialLeafMods{italic: styleOfficialBoolPtr(true)})...)
	want = append(want, styleOfficialMkLeaves(" not spooky", styleOfficialLeafMods{italic: styleOfficialBoolPtr(false)})...)
	styleOfficialCompareDecoration(t, src, got, want)
}

// --- Ported from ShadowColorTagTest.java --------------------------------

// styleOfficialShadowLeaf captures per-character shadow ARGB (nil = unset).
type styleOfficialShadowLeaf struct {
	ch    rune
	argb  uint32
	isSet bool
}

func styleOfficialFlattenShadow(comp c.Component) []styleOfficialShadowLeaf {
	var out []styleOfficialShadowLeaf
	if txt, ok := comp.(*c.Text); ok && txt.Content != "" {
		var l styleOfficialShadowLeaf
		if txt.S.ShadowColor != nil {
			l.isSet = true
			l.argb = txt.S.ShadowColor.ARGB
		}
		for _, r := range txt.Content {
			cp := l
			cp.ch = r
			out = append(out, cp)
		}
	}
	for _, e := range comp.Children() {
		out = append(out, styleOfficialFlattenShadow(e)...)
	}
	return out
}

func TestOfficialShadow_None(t *testing.T) {
	// testNoShadow: now i'm here <!shadow>and now i'm not! -> explicit "no shadow" (ARGB 0, isSet)
	root := deserialize(t, `now i'm here <!shadow>and now i'm not!`)
	got := styleOfficialFlattenShadow(root)
	idx := len([]rune("now i'm here "))
	for i, l := range got {
		wantSet := i >= idx
		if l.isSet != wantSet {
			t.Fatalf("char %d (%c): isSet=%v want %v", i, l.ch, l.isSet, wantSet)
		}
		if wantSet && l.argb != 0 {
			t.Fatalf("char %d (%c): argb=%#x want 0", i, l.ch, l.argb)
		}
	}
}

func TestOfficialShadow_NamedWithAlpha(t *testing.T) {
	// testRoundtripNamedShadow: <shadow:red:0.8>i have a red shadow -> shadowColor(RED, 0xCC)
	root := deserialize(t, `<shadow:red:0.8>i have a red shadow`)
	got := styleOfficialFlattenShadow(root)
	wantARGB := uint32(0xCC)<<24 | 0xff<<16 | 0x55<<8 | 0x55 // NamedTextColor.RED = #ff5555
	for i, l := range got {
		if !l.isSet || l.argb != wantARGB {
			t.Fatalf("char %d (%c): argb=%#x isSet=%v want %#x", i, l.ch, l.argb, l.isSet, wantARGB)
		}
	}
}

func TestOfficialShadow_HexWithAlpha(t *testing.T) {
	// testParseHexComponentShadow: <shadow:#FF0000:0.8>i have a redder shadow
	root := deserialize(t, `<shadow:#FF0000:0.8>i have a redder shadow`)
	got := styleOfficialFlattenShadow(root)
	wantARGB := uint32(0xCC)<<24 | 0xff<<16 | 0<<8 | 0
	for i, l := range got {
		if !l.isSet || l.argb != wantARGB {
			t.Fatalf("char %d (%c): argb=%#x isSet=%v want %#x", i, l.ch, l.argb, l.isSet, wantARGB)
		}
	}
}

func TestOfficialShadow_HexRGBA(t *testing.T) {
	// testSerializeShadow round-trip via #RRGGBBAA literal: <shadow:#054D79FF>
	root := deserialize(t, `<shadow:#054D79FF>This is a test`)
	got := styleOfficialFlattenShadow(root)
	wantARGB := uint32(0xFF)<<24 | 0x05<<16 | 0x4D<<8 | 0x79
	for i, l := range got {
		if !l.isSet || l.argb != wantARGB {
			t.Fatalf("char %d (%c): argb=%#x isSet=%v want %#x", i, l.ch, l.argb, l.isSet, wantARGB)
		}
	}
}

func TestOfficialShadow_Closing(t *testing.T) {
	// testSerializeShadowClosing: <shadow:#054D79FF>This is a</shadow> test
	root := deserialize(t, `<shadow:#054D79FF>This is a</shadow> test`)
	got := styleOfficialFlattenShadow(root)
	wantARGB := uint32(0xFF)<<24 | 0x05<<16 | 0x4D<<8 | 0x79
	idx := len([]rune("This is a"))
	for i, l := range got {
		wantSet := i < idx
		if l.isSet != wantSet {
			t.Fatalf("char %d (%c): isSet=%v want %v", i, l.ch, l.isSet, wantSet)
		}
		if wantSet && l.argb != wantARGB {
			t.Fatalf("char %d (%c): argb=%#x want %#x", i, l.ch, l.argb, wantARGB)
		}
	}
}

// --- Ported from FontTagTest.java ----------------------------------------

func styleOfficialFontOf(comp c.Component) []string {
	var out []string
	if txt, ok := comp.(*c.Text); ok && txt.Content != "" {
		f := ""
		if txt.S.Font != nil {
			f = txt.S.Font.String()
		}
		for range txt.Content {
			out = append(out, f)
		}
	}
	for _, e := range comp.Children() {
		out = append(out, styleOfficialFontOf(e)...)
	}
	return out
}

func TestOfficialFont_Namespaced(t *testing.T) {
	// testFont: Nothing <font:minecraft:uniform>Uniform <font:minecraft:alt>Alt  </font> Uniform
	src := `Nothing <font:minecraft:uniform>Uniform <font:minecraft:alt>Alt  </font> Uniform`
	root := deserialize(t, src)
	got := styleOfficialFontOf(root)
	want := []string{}
	want = append(want, repeatStr("", len("Nothing "))...)
	want = append(want, repeatStr("minecraft:uniform", len("Uniform "))...)
	want = append(want, repeatStr("minecraft:alt", len("Alt  "))...)
	want = append(want, repeatStr("minecraft:uniform", len(" Uniform"))...)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestOfficialFont_CustomNamespace(t *testing.T) {
	// testCustomFont: Default <font:myfont:best_font>Custom font <font:custom:worst_font>Another custom font </font>Back to previous font
	src := `Default <font:myfont:best_font>Custom font <font:custom:worst_font>Another custom font </font>Back to previous font`
	root := deserialize(t, src)
	got := styleOfficialFontOf(root)
	want := []string{}
	want = append(want, repeatStr("", len("Default "))...)
	want = append(want, repeatStr("myfont:best_font", len("Custom font "))...)
	want = append(want, repeatStr("custom:worst_font", len("Another custom font "))...)
	want = append(want, repeatStr("myfont:best_font", len("Back to previous font"))...)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestOfficialFont_NoNamespace(t *testing.T) {
	// testFontNoNamespace: Nothing <font:uniform>Uniform <font:alt>Alt  </font> Uniform
	src := `Nothing <font:uniform>Uniform <font:alt>Alt  </font> Uniform`
	root := deserialize(t, src)
	got := styleOfficialFontOf(root)
	want := []string{}
	want = append(want, repeatStr("", len("Nothing "))...)
	want = append(want, repeatStr("minecraft:uniform", len("Uniform "))...)
	want = append(want, repeatStr("minecraft:alt", len("Alt  "))...)
	want = append(want, repeatStr("minecraft:uniform", len(" Uniform"))...)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func repeatStr(s string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}

// --- Ported from InsertionTagTest.java -----------------------------------

func TestOfficialInsertion_Basic(t *testing.T) {
	// testInsertion: Click <insert:test>this</insert> to insert!
	root := deserialize(t, `Click <insert:test>this</insert> to insert!`)
	got := styleOfficialInsertionOf(root)
	want := []string{}
	want = append(want, repeatStr("", len("Click "))...)
	want = append(want, repeatStr("test", len("this"))...)
	want = append(want, repeatStr("", len(" to insert!"))...)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func styleOfficialInsertionOf(comp c.Component) []string {
	var out []string
	if txt, ok := comp.(*c.Text); ok && txt.Content != "" {
		ins := ""
		if txt.S.Insertion != nil {
			ins = *txt.S.Insertion
		}
		for range txt.Content {
			out = append(out, ins)
		}
	}
	for _, e := range comp.Children() {
		out = append(out, styleOfficialInsertionOf(e)...)
	}
	return out
}
