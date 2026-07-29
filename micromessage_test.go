package micromessage

import "testing"

func assertPlain(t *testing.T, input, expected string) {
	t.Helper()
	c, err := Deserialize(input)
	if err != nil {
		t.Fatalf("Deserialize(%q) error: %v", input, err)
	}
	got := PlainText(c)
	if got != expected {
		t.Errorf("Deserialize(%q).PlainText() = %q, want %q", input, got, expected)
	}
}

func TestBasicColor(t *testing.T) {
	assertPlain(t, "<red>Hello, <b>World</b></red>", "Hello, World")
}

func TestColorSimple(t *testing.T) {
	c, err := Deserialize("<yellow>TEST")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(c.Children))
	}
	child := c.Children[0]
	if child.Style.Color == nil || NamedColorName(child.Style.Color) != "yellow" {
		t.Errorf("expected yellow color, got %+v", child.Style.Color)
	}
	if PlainText(c) != "TEST" {
		t.Errorf("got %q", PlainText(c))
	}
}

func TestNestedColor(t *testing.T) {
	input := "<yellow>TEST<green> nested</green>Test"
	assertPlain(t, input, "TEST nestedTest")
}

func TestColorAliases(t *testing.T) {
	for _, pair := range [][2]string{
		{"<grey>This is english", "<gray>This is english"},
		{"<dark_grey>This is still english", "<dark_gray>This is still english"},
		{"<colour:grey>This is english", "<color:gray>This is english"},
	} {
		a, err := Deserialize(pair[0])
		if err != nil {
			t.Fatal(err)
		}
		b, err := Deserialize(pair[1])
		if err != nil {
			t.Fatal(err)
		}
		if a.Children[0].Style.Color.HexString() != b.Children[0].Style.Color.HexString() {
			t.Errorf("%q and %q resolved to different colors", pair[0], pair[1])
		}
	}
}

func TestHexColor(t *testing.T) {
	c, err := Deserialize("<color:#ff00ff>TEST")
	if err != nil {
		t.Fatal(err)
	}
	if c.Children[0].Style.Color.HexString() != "#FF00FF" {
		t.Errorf("got %s", c.Children[0].Style.Color.HexString())
	}
}

func TestHexColorShort(t *testing.T) {
	c, err := Deserialize("<#ff00ff>TEST")
	if err != nil {
		t.Fatal(err)
	}
	if c.Children[0].Style.Color.HexString() != "#FF00FF" {
		t.Errorf("got %s", c.Children[0].Style.Color.HexString())
	}
}

func TestDecorationShorthand(t *testing.T) {
	c, err := Deserialize("<b>bold</b> normal")
	if err != nil {
		t.Fatal(err)
	}
	if c.Children[0].Style.Decorations[Bold] != True {
		t.Errorf("expected bold true")
	}
	if PlainText(c) != "bold normal" {
		t.Errorf("got %q", PlainText(c))
	}
}

func TestDisabledDecoration(t *testing.T) {
	input := "<italic:false>Test<bold:false>Test2<bold>Test3"
	c, err := Deserialize(input)
	if err != nil {
		t.Fatal(err)
	}
	if c.Children[0].Style.Decorations[Italic] != False {
		t.Errorf("expected italic false at root")
	}
	if PlainText(c) != "TestTest2Test3" {
		t.Errorf("got %q", PlainText(c))
	}
}

func TestDisabledDecorationShorthand(t *testing.T) {
	input := "<!italic>Test<!bold>Test2<bold>Test3"
	c, err := Deserialize(input)
	if err != nil {
		t.Fatal(err)
	}
	if c.Children[0].Style.Decorations[Italic] != False {
		t.Errorf("expected italic false at root")
	}
}

func TestInvalidTag(t *testing.T) {
	input := "<red><test>"
	c, err := Deserialize(input)
	if err != nil {
		t.Fatal(err)
	}
	if PlainText(c) != "<test>" {
		t.Errorf("got %q", PlainText(c))
	}
}

func TestStripSimple(t *testing.T) {
	input := "<yellow>TEST<green> nested</green>Test"
	if got := StripTags(input); got != "TEST nestedTest" {
		t.Errorf("got %q", got)
	}
}

func TestStripComplex(t *testing.T) {
	input := "<yellow><test> random <bold>stranger</bold><click:run_command:test command><underlined><red>click here</click><blue> to <bold>FEEL</underlined> it"
	expected := "<test> random strangerclick here to FEEL it"
	if got := StripTags(input); got != expected {
		t.Errorf("got %q want %q", got, expected)
	}
}

func TestStripTagsWithPlaceholder(t *testing.T) {
	input := "Hello, <red><name>!"
	got := StripTags(input, StringPlaceholder("name", "you"))
	if got != "Hello, !" {
		t.Errorf("got %q", got)
	}
}

func TestPlaceholderComponent(t *testing.T) {
	input := "<yellow><test> random"
	c, err := Deserialize(input, ComponentPlaceholder("test", Text("Hello!")))
	if err != nil {
		t.Fatal(err)
	}
	if PlainText(c) != "Hello! random" {
		t.Errorf("got %q", PlainText(c))
	}
}

func TestClickTag(t *testing.T) {
	c, err := Deserialize("<click:run_command:test command>click here</click> after")
	if err != nil {
		t.Fatal(err)
	}
	if c.Children[0].Style.ClickEvent == nil || c.Children[0].Style.ClickEvent.Action != RunCommand {
		t.Fatalf("expected run_command click event, got %+v", c.Children[0].Style.ClickEvent)
	}
	if c.Children[0].Style.ClickEvent.Value != "test command" {
		t.Errorf("got %q", c.Children[0].Style.ClickEvent.Value)
	}
}

func TestHoverShowText(t *testing.T) {
	c, err := Deserialize("<hover:show_text:'This is a test'>hovered")
	if err != nil {
		t.Fatal(err)
	}
	hv := c.Children[0].Style.HoverEvent
	if hv == nil || hv.Action != ShowText {
		t.Fatalf("expected show_text hover event")
	}
	if PlainText(hv.Value) != "This is a test" {
		t.Errorf("got %q", PlainText(hv.Value))
	}
}

func TestInsertionTag(t *testing.T) {
	c, err := Deserialize("<insert:test>text")
	if err != nil {
		t.Fatal(err)
	}
	ins := c.Children[0].Style.Insertion
	if ins == nil || *ins != "test" {
		t.Errorf("got %v", ins)
	}
}

func TestFontTag(t *testing.T) {
	c, err := Deserialize("<font:minecraft:alt>text")
	if err != nil {
		t.Fatal(err)
	}
	f := c.Children[0].Style.Font
	if f == nil || *f != "minecraft:alt" {
		t.Errorf("got %v", f)
	}
}

func TestResetTag(t *testing.T) {
	input := "<red><reset>plain"
	c, err := Deserialize(input)
	if err != nil {
		t.Fatal(err)
	}
	if PlainText(c) != "plain" {
		t.Errorf("got %q", PlainText(c))
	}
}

func TestNewlineTag(t *testing.T) {
	c, err := Deserialize("line1<newline>line2<br>line3")
	if err != nil {
		t.Fatal(err)
	}
	if PlainText(c) != "line1\nline2\nline3" {
		t.Errorf("got %q", PlainText(c))
	}
}

func TestGradient(t *testing.T) {
	input := "<gradient:red:blue>AB</gradient>"
	c, err := Deserialize(input)
	if err != nil {
		t.Fatal(err)
	}
	if PlainText(c) != "AB" {
		t.Errorf("got %q", PlainText(c))
	}
	// first char should be pure red, last char pure blue
	first := findFirstLeafColor(c)
	if first == nil || first.HexString() != "#FF5555" {
		t.Errorf("expected first color red (#FF5555), got %v", first)
	}
}

func TestRainbow(t *testing.T) {
	input := "<rainbow>||||||||||||||||||||||||</rainbow>!"
	c, err := Deserialize(input)
	if err != nil {
		t.Fatal(err)
	}
	if PlainText(c) != "||||||||||||||||||||||||!" {
		t.Errorf("got %q", PlainText(c))
	}
}

func TestEscaping(t *testing.T) {
	input := "\\<yellow>not colored"
	c, err := Deserialize(input)
	if err != nil {
		t.Fatal(err)
	}
	if PlainText(c) != "<yellow>not colored" {
		t.Errorf("got %q", PlainText(c))
	}
}

func TestEscapeTags(t *testing.T) {
	input := "<yellow>TEST<green> nested</green>Test"
	expected := "\\<yellow>TEST\\<green> nested\\</green>Test"
	if got := EscapeTags(input); got != expected {
		t.Errorf("got %q want %q", got, expected)
	}
}

func TestSerializeRoundtrip(t *testing.T) {
	input := "<red>This is a test"
	c, err := Deserialize(input)
	if err != nil {
		t.Fatal(err)
	}
	out := Serialize(c)
	if out != input {
		t.Errorf("got %q want %q", out, input)
	}
}

func TestSerializeDecoration(t *testing.T) {
	c := Empty()
	inner := Empty()
	inner.Style.Decorations = map[string]TriState{Underlined: True}
	bold := Text("underlined")
	bold.Style.Decorations = map[string]TriState{Bold: True}
	inner.Append(bold)
	c.Append(inner, Text(", this isn't"))

	expected := "<underlined><bold>underlined</bold></underlined>, this isn't"
	if got := Serialize(c); got != expected {
		t.Errorf("got %q want %q", got, expected)
	}
}

func findFirstLeafColor(c *Component) *Color {
	if c.Text != "" {
		return c.Style.Color
	}
	for _, ch := range c.Children {
		if col := findFirstLeafColor(ch); col != nil {
			return col
		}
	}
	return nil
}
