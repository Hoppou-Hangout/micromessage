package micromessage

import (
	"testing"

	c "go.minekube.com/common/minecraft/component"
)

// --- Gradient: extra cases from Adventure's GradientTagTest.java not ------
// --- already covered by gradient_test.go -----------------------------------

func TestGradientOfficial_WithHover(t *testing.T) {
	hexes := []string{
		"#ffffff", "#f4f4f4", "#e9e9e9", "#dedede", "#d3d3d3", "#c8c8c8", "#bcbcbc", "#b1b1b1",
		"#a6a6a6", "#9b9b9b", "#909090", "#858585", "#7a7a7a", "#6f6f6f", "#646464", "#595959",
		"#4e4e4e", "#434343", "#373737", "#2c2c2c", "#212121", "#161616", "#0b0b0b", "#000000",
	}
	body := grad('|', hexes...)
	excl := text("!", "#ffff55")
	const hover = "show_text:This is a test"
	for i := range body {
		body[i].hover = hover
	}
	for i := range excl {
		excl[i].hover = hover
	}
	check(t, "gradient_with_hover",
		`<yellow>Woo: <hover:show_text:'This is a test'><gradient>||||||||||||||||||||||||</gradient>!`,
		cat(text("Woo: ", "#ffff55"), body, excl))
}

func TestGradientOfficial_NestedInnerGradientWithInnerBold(t *testing.T) {
	check(t, "nested_inner_gradient_with_inner_token",
		`<gradient:green:blue>123<gradient:red:yellow>456<bold>789</gradient>abc</gradient>!`,
		cat(
			[]charLeaf{{ch: '1', color: "#55ff55"}, {ch: '2', color: "#55f064"}, {ch: '3', color: "#55e074"}},
			[]charLeaf{{ch: '4', color: "#ff5555"}, {ch: '5', color: "#ff7755"}, {ch: '6', color: "#ff9955"}},
			[]charLeaf{
				{ch: '7', color: "#ffbb55", bold: true},
				{ch: '8', color: "#ffdd55", bold: true},
				{ch: '9', color: "#ffff55", bold: true},
			},
			[]charLeaf{{ch: 'a', color: "#5574e0"}, {ch: 'b', color: "#5564f0"}, {ch: 'c', color: "#5555ff"}},
			text("!", ""),
		))
}

func TestGradientOfficial_DecorationsPreserved(t *testing.T) {
	// GH-790: a placeholder component's own decorations (italic here) survive
	// gradient coloring instead of being clobbered.
	placeholder := &c.Text{Content: "b", S: c.Style{Italic: c.True}}
	root := deserialize(t, `<gradient>a<placeholder/>c<bold>d</bold>!</gradient>`,
		WithTagResolver(Placeholder("placeholder", ComponentTag(placeholder))))
	checkLeaves(t, flattenComponent(root), cat(
		text("a", "#ffffff"),
		[]charLeaf{{ch: 'b', color: "#bfbfbf", italic: true}},
		text("c", "#808080"),
		textB("d", "#404040"),
		text("!", "#000000"),
	))
}

func TestGradientOfficial_LangTagInGradient(t *testing.T) {
	// GH-827: a <lang>/<tr> tag inside a gradient still gets its own gradient
	// stop color, and gradient indexing continues through it correctly for
	// surrounding text.
	root := deserialize(t, `<gradient:red:blue>ab<lang:block.minecraft.diamond_block>!</gradient>`)
	checkLeaves(t, flattenComponent(root), cat(
		text("a", "#ff5555"),
		[]charLeaf{{ch: 'b', color: "#c6558e"}},
		text("!", "#5555ff"),
	))

	var tr *c.Translation
	var find func(comp c.Component)
	find = func(comp c.Component) {
		if t2, ok := comp.(*c.Translation); ok {
			tr = t2
		}
		for _, ch := range comp.Children() {
			find(ch)
		}
	}
	find(root)
	if tr == nil {
		t.Fatalf("no translation component found")
	}
	if tr.Key != "block.minecraft.diamond_block" {
		t.Fatalf("unexpected translation key %q", tr.Key)
	}
	if tr.S.Color == nil || tr.S.Color.Hex() != "#8e55c6" {
		got := ""
		if tr.S.Color != nil {
			got = tr.S.Color.Hex()
		}
		t.Fatalf("translation color = %q, want #8e55c6", got)
	}
}

func TestGradientOfficial_Gh137(t *testing.T) {
	const input1 = `<gradient:gold:yellow:red><dum>`
	const input2 = `<gradient:gold:yellow:red><dum>a`

	dumComp := func(s string) TagResolver {
		return Placeholder("dum", ComponentTag(&c.Text{Content: s}))
	}

	cases := []struct {
		name string
		src  string
		dum  string
		want []charLeaf
	}{
		{"1_char", input1, "a", []charLeaf{{ch: 'a', color: "#ffaa00"}}},
		{"2_char", input1, "aa", []charLeaf{{ch: 'a', color: "#ffaa00"}, {ch: 'a', color: "#ff5555"}}},
		{"3_char", input1, "aaa", []charLeaf{
			{ch: 'a', color: "#ffaa00"}, {ch: 'a', color: "#ffff55"}, {ch: 'a', color: "#ff5555"},
		}},
		{"4_char", input1, "aaaa", []charLeaf{
			{ch: 'a', color: "#ffaa00"}, {ch: 'a', color: "#ffe339"}, {ch: 'a', color: "#ffc655"}, {ch: 'a', color: "#ff5555"},
		}},
		{"3_char_plus_trailing", input2, "aaa", []charLeaf{
			{ch: 'a', color: "#ffaa00"}, {ch: 'a', color: "#ffe339"}, {ch: 'a', color: "#ffc655"}, {ch: 'a', color: "#ff5555"},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := deserialize(t, tc.src, WithTagResolver(dumComp(tc.dum)))
			checkLeaves(t, flattenComponent(root), tc.want)
		})
	}
}

// --- Rainbow: ported from Adventure's RainbowTagTest.java (no existing ----
// --- Go coverage) -----------------------------------------------------------

func TestRainbowOfficial_Default(t *testing.T) {
	check(t, "rainbow", `<yellow>Woo: <rainbow>||||||||||||||||||||||||</rainbow>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|', "#ff0000", "#ff3f00", "#ff7f00", "#ffbf00", "#ffff00", "#bfff00", "#7fff00", "#3fff00",
				"#00ff00", "#00ff3f", "#00ff7f", "#00ffbf", "#00ffff", "#00bfff", "#007fff", "#003fff",
				"#0000ff", "#3f00ff", "#7f00ff", "#bf00ff", "#ff00ff", "#ff00bf", "#ff007f", "#ff003f"),
			text("!", "#ffff55"),
		))
}

func TestRainbowOfficial_Backwards(t *testing.T) {
	check(t, "rainbow_backwards", `<yellow>Woo: <rainbow:!>||||||||||||||||||||||||</rainbow>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|', "#ff003f", "#ff007f", "#ff00bf", "#ff00ff", "#bf00ff", "#7f00ff", "#3f00ff", "#0000ff",
				"#003fff", "#007fff", "#00bfff", "#00ffff", "#00ffbf", "#00ff7f", "#00ff3f", "#00ff00",
				"#3fff00", "#7fff00", "#bfff00", "#ffff00", "#ffbf00", "#ff7f00", "#ff3f00", "#ff0000"),
			text("!", "#ffff55"),
		))
}

func TestRainbowOfficial_Phase(t *testing.T) {
	check(t, "rainbow_phase_2", `<yellow>Woo: <rainbow:2>||||||||||||||||||||||||</rainbow>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|', "#cbff00", "#8cff00", "#4cff00", "#0cff00", "#00ff33", "#00ff72", "#00ffb2", "#00fff2",
				"#00cbff", "#008cff", "#004cff", "#000cff", "#3200ff", "#7200ff", "#b200ff", "#f200ff",
				"#ff00cc", "#ff008c", "#ff004c", "#ff000c", "#ff3200", "#ff7200", "#ffb200", "#fff200"),
			text("!", "#ffff55"),
		))
}

func TestRainbowOfficial_PhaseBackwards(t *testing.T) {
	check(t, "rainbow_phase_backwards_2", `<yellow>Woo: <rainbow:!2>||||||||||||||||||||||||</rainbow>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|', "#fff200", "#ffb200", "#ff7200", "#ff3200", "#ff000c", "#ff004c", "#ff008c", "#ff00cc",
				"#f200ff", "#b200ff", "#7200ff", "#3200ff", "#000cff", "#004cff", "#008cff", "#00cbff",
				"#00fff2", "#00ffb2", "#00ff72", "#00ff33", "#0cff00", "#4cff00", "#8cff00", "#cbff00"),
			text("!", "#ffff55"),
		))
}

func TestRainbowOfficial_WithInsertion(t *testing.T) {
	check(t, "rainbow_with_insertion", `<yellow>Woo: <insert:test><rainbow>||||||||||||||||||||||||</rainbow>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|', "#ff0000", "#ff3f00", "#ff7f00", "#ffbf00", "#ffff00", "#bfff00", "#7fff00", "#3fff00",
				"#00ff00", "#00ff3f", "#00ff7f", "#00ffbf", "#00ffff", "#00bfff", "#007fff", "#003fff",
				"#0000ff", "#3f00ff", "#7f00ff", "#bf00ff", "#ff00ff", "#ff00bf", "#ff007f", "#ff003f"),
			text("!", "#ffff55"),
		))
}

func TestRainbowOfficial_BackwardsWithInsertion(t *testing.T) {
	check(t, "rainbow_backwards_with_insertion", `<yellow>Woo: <insert:test><rainbow:!>||||||||||||||||||||||||</rainbow>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|', "#ff003f", "#ff007f", "#ff00bf", "#ff00ff", "#bf00ff", "#7f00ff", "#3f00ff", "#0000ff",
				"#003fff", "#007fff", "#00bfff", "#00ffff", "#00ffbf", "#00ff7f", "#00ff3f", "#00ff00",
				"#3fff00", "#7fff00", "#bfff00", "#ffff00", "#ffbf00", "#ff7f00", "#ff3f00", "#ff0000"),
			text("!", "#ffff55"),
		))
}

func TestRainbowOfficial_WithInnerClick(t *testing.T) {
	check(t, "rainbow_with_inner_click",
		`Rainbow: <rainbow><click:open_url:'https://github.com'>GH</click></rainbow>`,
		cat(
			text("Rainbow: ", ""),
			[]charLeaf{
				{ch: 'G', color: "#ff0000", click: "open_url:https://github.com"},
				{ch: 'H', color: "#00ffff", click: "open_url:https://github.com"},
			},
		))
}

func TestRainbowOfficial_BackwardsWithInnerClick(t *testing.T) {
	check(t, "rainbow_backwards_with_inner_click",
		`Rainbow: <rainbow:!0><click:open_url:'https://github.com'>GH</click></rainbow>`,
		cat(
			text("Rainbow: ", ""),
			[]charLeaf{
				{ch: 'G', color: "#00ffff", click: "open_url:https://github.com"},
				{ch: 'H', color: "#ff0000", click: "open_url:https://github.com"},
			},
		))
}

func TestRainbowOfficial_NoRepeatedTextAfterUnclosed(t *testing.T) {
	// GH-125
	check(t, "gh125", `<rainbow>rainbow<yellow>yellow`,
		cat(
			gradRunes("rainbow", "#ff0000", "#ff7500", "#ffeb00", "#9cff00", "#27ff00", "#00ff4e", "#00ffc4"),
			text("yellow", "#ffff55"),
		))
}

func TestRainbowOfficial_ContinuesAfterColoredInner(t *testing.T) {
	check(t, "rainbow_continues_after_colored_inner",
		`<rainbow>rain<white>white</white>bow`,
		cat(
			gradRunes("rain", "#ff0000", "#ff7f00", "#ffff00", "#7fff00"),
			text("white", "#ffffff"),
			gradRunes("bow", "#7f00ff", "#ff00ff", "#ff007f"),
		))

	check(t, "gradient_continues_after_colored_inner",
		`<gradient>grad<green>green</green>ient`,
		cat(
			gradRunes("grad", "#ffffff", "#eaeaea", "#d5d5d5", "#bfbfbf"),
			text("green", "#55ff55"),
			gradRunes("ient", "#404040", "#2b2b2b", "#151515", "#000000"),
		))
}

func TestRainbowOfficial_Gh147(t *testing.T) {
	root := deserialize(t, `<rainbow><msg>`, WithTagResolver(Placeholder("msg", ComponentTag(&c.Text{Content: "yo"}))))
	checkLeaves(t, flattenComponent(root), []charLeaf{
		{ch: 'y', color: "#ff0000"},
		{ch: 'o', color: "#00ffff"},
	})
}

func TestRainbowOfficial_Gh1040(t *testing.T) {
	check(t, "gh1040", `<rainbow:16777215>||||||||||`,
		grad('|', "#00ffff", "#0065ff", "#3200ff", "#cc00ff", "#ff0099",
			"#ff0000", "#ff9900", "#ccff00", "#32ff00", "#00ff65"))
}

// --- Transition: ported from Adventure's TransitionTagTest.java (no ---------
// --- existing Go coverage) ---------------------------------------------------

func TestTransitionOfficial_ColorAtPhase(t *testing.T) {
	cases := []struct {
		phase string
		color string
	}{
		{"-1.0", "#000000"},
		{"0.5", "#808080"},
		{"0.0", "#ffffff"},
		{"0.5", "#808080"},
		{"1.0", "#000000"},
	}
	for _, tc := range cases {
		check(t, "phase_"+tc.phase, `<transition:white:black:`+tc.phase+`>Hello World`,
			text("Hello World", tc.color))
	}
}

// --- Reset: ported from Adventure's ResetTagTest.java (no existing Go -------
// --- coverage) ----------------------------------------------------------------

func TestResetOfficial_Basic(t *testing.T) {
	check(t, "reset", `Click <yellow><insert:test>this<rainbow> wooo<reset> to insert!`,
		cat(
			text("Click ", ""),
			text("this", "#ffff55"),
			gradRunes(" wooo", "#ff0000", "#cbff00", "#00ff66", "#0065ff", "#cc00ff"),
			text(" to insert!", ""),
		))
}

// --- Newline: ported from Adventure's NewlineTagTest.java (no existing Go ---
// --- coverage) ------------------------------------------------------------------

func TestNewlineOfficial_Basic(t *testing.T) {
	check(t, "newline", `<red>Line<br><gray>break!`,
		cat(
			text("Line\n", "#ff5555"),
			text("break!", "#aaaaaa"),
		))
}
