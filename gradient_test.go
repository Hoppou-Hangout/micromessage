package micromessage

import "testing"

func TestGradient_Default(t *testing.T) {
	// testGradient: <gradient> with no args defaults to white -> black.
	check(t, "default_white_black", `<yellow>Woo: <gradient>||||||||||||||||||||||||</gradient>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|', "#ffffff", "#f4f4f4", "#e9e9e9", "#dedede", "#d3d3d3", "#c8c8c8", "#bcbcbc", "#b1b1b1",
				"#a6a6a6", "#9b9b9b", "#909090", "#858585", "#7a7a7a", "#6f6f6f", "#646464", "#595959",
				"#4e4e4e", "#434343", "#373737", "#2c2c2c", "#212121", "#161616", "#0b0b0b", "#000000"),
			text("!", "#ffff55"),
		))
}

func TestGradient_Hex2Color(t *testing.T) {
	check(t, "hex_5e4fa2_f79459", `<yellow>Woo: <gradient:#5e4fa2:#f79459>||||||||||||||||||||||||</gradient>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|', "#5e4fa2", "#65529f", "#6b559c", "#725898", "#795b95", "#7f5e92", "#86618f", "#8d648c",
				"#936789", "#9a6a85", "#a16d82", "#a7707f", "#ae737c", "#b47679", "#bb7976", "#c27c72",
				"#c87f6f", "#cf826c", "#d68569", "#dc8866", "#e38b63", "#ea8e5f", "#f0915c", "#f79459"),
			text("!", "#ffff55"),
		))
}

func TestGradient_NamedGreenBlue(t *testing.T) {
	check(t, "green_blue", `<yellow>Woo: <gradient:green:blue>||||||||||||||||||||||||</gradient>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|', "#55ff55", "#55f85c", "#55f064", "#55e96b", "#55e173", "#55da7a", "#55d381", "#55cb89",
				"#55c490", "#55bc98", "#55b59f", "#55aea6", "#55a6ae", "#559fb5", "#5598bc", "#5590c4",
				"#5589cb", "#5581d3", "#557ada", "#5573e1", "#556be9", "#5564f0", "#555cf8", "#5555ff"),
			text("!", "#ffff55"),
		))
}

func TestGradient_Phase(t *testing.T) {
	check(t, "green_blue_phase_0.7", `<yellow>Woo: <gradient:green:blue:0.7>||||||||||||||||||||||||</gradient>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|', "#5588cc", "#5581d3", "#5579db", "#5572e2", "#556aea", "#5563f1", "#555cf8", "#5556fe",
				"#555df7", "#5565ef", "#556ce8", "#5573e1", "#557bd9", "#5582d2", "#5589cb", "#5591c3",
				"#5598bc", "#55a0b4", "#55a7ad", "#55aea6", "#55b69e", "#55bd97", "#55c58f", "#55cc88"),
			text("!", "#ffff55"),
		))
}

func TestGradient_MultiColor(t *testing.T) {
	check(t, "5_stop", `<yellow>Woo: <gradient:red:blue:green:yellow:red>||||||||||||||||||||||||||||||||||||||||||||||||||||||</gradient>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|',
				"#ff5555", "#f25562", "#e5556f", "#d9557b", "#cc5588", "#bf5595", "#b255a2", "#a555af",
				"#9855bc", "#8c55c8", "#7f55d5", "#7255e2", "#6555ef", "#5855fc", "#555ff5", "#556be9",
				"#5578dc", "#5585cf", "#5592c2", "#559fb5", "#55aca8", "#55b89c", "#55c58f", "#55d282",
				"#55df75", "#55ec68", "#55f95b", "#5bff55", "#68ff55", "#75ff55", "#82ff55", "#8fff55",
				"#9cff55", "#a8ff55", "#b5ff55", "#c2ff55", "#cfff55", "#dcff55", "#e9ff55", "#f5ff55",
				"#fffc55", "#ffef55", "#ffe255", "#ffd555", "#ffc855", "#ffbc55", "#ffaf55", "#ffa255",
				"#ff9555", "#ff8855", "#ff7b55", "#ff6f55", "#ff6255", "#ff5555"),
			text("!", "#ffff55"),
		))
}

func TestGradient_MultiColor2_BlackWhiteBlack(t *testing.T) {
	check(t, "black_white_black", `<yellow>Woo: <gradient:black:white:black>||||||||||||||||||||||||||||||||||||||||||||||||||||||</gradient>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|',
				"#000000", "#0a0a0a", "#131313", "#1d1d1d", "#262626", "#303030", "#3a3a3a", "#434343",
				"#4d4d4d", "#575757", "#606060", "#6a6a6a", "#737373", "#7d7d7d", "#878787", "#909090",
				"#9a9a9a", "#a4a4a4", "#adadad", "#b7b7b7", "#c0c0c0", "#cacaca", "#d4d4d4", "#dddddd",
				"#e7e7e7", "#f1f1f1", "#fafafa", "#fafafa", "#f1f1f1", "#e7e7e7", "#dddddd", "#d4d4d4",
				"#cacaca", "#c0c0c0", "#b7b7b7", "#adadad", "#a4a4a4", "#9a9a9a", "#909090", "#878787",
				"#7d7d7d", "#737373", "#6a6a6a", "#606060", "#575757", "#4d4d4d", "#434343", "#3a3a3a",
				"#303030", "#262626", "#1d1d1d", "#131313", "#0a0a0a", "#000000"),
			text("!", "#ffff55"),
		))
}

func TestGradient_MultiColor2_NegativePhase(t *testing.T) {
	check(t, "black_white_black_phase_-0.65", `<yellow>Woo: <gradient:black:white:black:-0.65>||||||||||||||||||||||||||||||||||||||||||||||||||||||</gradient>!`,
		cat(
			text("Woo: ", "#ffff55"),
			grad('|',
				"#b3b3b3", "#bcbcbc", "#c6c6c6", "#cfcfcf", "#d9d9d9", "#e3e3e3", "#ececec", "#f6f6f6",
				"#ffffff", "#f5f5f5", "#ebebeb", "#e2e2e2", "#d8d8d8", "#cecece", "#c5c5c5", "#bbbbbb",
				"#b2b2b2", "#a8a8a8", "#9e9e9e", "#959595", "#8b8b8b", "#818181", "#787878", "#6e6e6e",
				"#656565", "#5b5b5b", "#515151", "#484848", "#3e3e3e", "#343434", "#2b2b2b", "#212121",
				"#181818", "#0e0e0e", "#040404", "#000000", "#000000", "#000000", "#000000", "#000000",
				"#000000", "#000000", "#000000", "#000000", "#000000", "#000000", "#000000", "#000000",
				"#000000", "#000000", "#000000", "#000000", "#000000", "#000000"),
			text("!", "#ffff55"),
		))
}

func TestGradient_MultiColorPhased(t *testing.T) {
	check(t, "3_stop_phase_1", `<gradient:#aa0000:#ff5555:#aaaaaa:1>-------`,
		grad('-', "#aaaaaa", "#aa7171", "#aa3939", "#aa0000", "#c61c1c", "#e33939", "#ff5555"))
}

func TestGradient_NonBMPCharacters(t *testing.T) {
	// Confirms gradient length/index counts by rune (codepoint), not UTF-16
	// code units - relevant since these are non-BMP surrogate-pair chars.
	check(t, "non_bmp_codepoints", `Something <gradient:green:blue:1.0>𐌰𐌱𐌲</gradient>`,
		cat(
			text("Something ", ""),
			gradRunes("𐌰𐌱𐌲", "#5555ff", "#55aaaa", "#55ff55"),
		))
}

func TestGradient_SingleChar(t *testing.T) {
	check(t, "single_char_3_stop", `<gradient:red:blue:green>A`, grad('A', "#ff5555"))
}

func TestGradient_SingleCharMultiChar(t *testing.T) {
	// "AB" is still measured as size=2, but only the first char is what the
	// java test actually names "single char" - this covers both entries.
	got := renderSrc(t, `<gradient:red:blue:green:red>AB`)
	if len(got) != 2 || got[0].ch != 'A' || got[0].color != "#ff5555" {
		t.Fatalf("unexpected: %s", fmtLeaves(got))
	}
}

func TestGradient_WithInnerBold(t *testing.T) {
	check(t, "inner_bold", `<gradient:green:blue>123<bold>456</gradient>!`,
		cat(
			grad('1', "#55ff55"),
			[]charLeaf{{ch: '2', color: "#55dd77"}, {ch: '3', color: "#55bb99"}},
			[]charLeaf{
				{ch: '4', color: "#5599bb", bold: true},
				{ch: '5', color: "#5577dd", bold: true},
				{ch: '6', color: "#5555ff", bold: true},
			},
			text("!", ""),
		))
}

func TestGradient_NestedDoesNotOverrideOuter(t *testing.T) {
	// GH-510: a nested gradient's own characters get the nested gradient's
	// colors, but the outer gradient's index accounting still advances through
	// them, so surrounding chars land on the right outer color.
	check(t, "gh510_a", `<gradient:#1985ff:#2bc7ff>a<gradient:#00fffb:#00ffc3>b</gradient> <gray>gray</gray></gradient>`,
		cat(
			[]charLeaf{{ch: 'a', color: "#1985ff"}},
			[]charLeaf{{ch: 'b', color: "#00fffb"}},
			[]charLeaf{{ch: ' ', color: "#1f9bff"}},
			text("gray", "#aaaaaa"),
		))
}

func TestGradient_NestedReallyDoesNotOverride(t *testing.T) {
	check(t, "gh510_b", `<gradient:white:blue>A <gradient:yellow:black>B <white>C`,
		cat(
			[]charLeaf{{ch: 'A', color: "#ffffff"}, {ch: ' ', color: "#d5d5ff"}},
			[]charLeaf{{ch: 'B', color: "#ffff55"}, {ch: ' ', color: "#80802b"}},
			[]charLeaf{{ch: 'C', color: "#ffffff"}},
		))
}
