package micromessage

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

// --- Lexer -----------------------------------------------------------------
//
// MiniMessage tag name and arguments are separate things: <tag:arg1:arg2>.
// Arguments may be bare words or quoted strings (single or double quotes, with
// backslash escapes), and bare words may themselves contain slashes (e.g.
// <click:run_command:/help>), so we can't just glob the whole thing into one
// identifier like the previous version did.

var mmLexer = lexer.MustStateful(lexer.Rules{
	"Root": {
		{Name: "TagOpen", Pattern: `<`, Action: lexer.Push("Tag")},
		{Name: "Text", Pattern: `[^<]+`},
	},
	"Tag": {
		{Name: "SelfClose", Pattern: `/>`, Action: lexer.Pop()},
		{Name: "Close", Pattern: `>`, Action: lexer.Pop()},
		{Name: "Colon", Pattern: `:`},
		{Name: "String", Pattern: `"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'`},
		{Name: "Ident", Pattern: `[!?#]?[a-zA-Z0-9_][\w.-]*`},
		{Name: "Arg", Pattern: `[^:<>'"]+`},
	},
})

// --- AST ---------------------------------------------------------------

type Kind int

const (
	KindText Kind = iota
	KindElement
)

// Node is either a run of text or a tag, mirroring how MiniMessage really
// works: styling tags don't have "children" in a strict sense, they wrap
// whatever text/tags follow until an explicit close or the end of input.
type Node struct {
	Kind Kind

	Text string // set when Kind == KindText

	Name       string // tag name, set when Kind == KindElement
	Args       []string
	SelfClosed bool // true for <tag/>
	Closed     bool // true if an explicit </tag> was found; false = auto-closed
	Children   []*Node
}

// resetTag never takes children or a close tag: <reset> just resets state.
const resetTag = "reset"

// leafTags never take children or a close tag, same as resetTag: they insert
// a single node (a newline, a translation component, ...) rather than
// wrapping content.
var leafTags = map[string]bool{
	"br": true, "newline": true,
	"lang": true, "tr": true, "translate": true,
	"lang_or": true, "tr_or": true, "translate_or": true,
}

// --- Parser --------------------------------------------------------------

type parser struct {
	lex *lexer.PeekingLexer
	sym map[string]lexer.TokenType
}

func newParser(src string) (*parser, error) {
	lx, err := mmLexer.LexString("", src)
	if err != nil {
		return nil, err
	}
	sym := mmLexer.Symbols()
	pl, err := lexer.Upgrade(lx)
	if err != nil {
		return nil, err
	}
	return &parser{lex: pl, sym: sym}, nil
}

func (p *parser) tt(name string) lexer.TokenType { return p.sym[name] }

// Parse consumes the entire input and returns the top-level nodes.
func (p *parser) Parse() ([]*Node, error) {
	nodes, _, err := p.parseContentUntil("")
	return nodes, err
}

// parseContentUntil consumes text and elements until it sees a closing tag
// matching closeName, hits EOF, or (only at top level, closeName=="") skips
// over a stray closing tag that doesn't belong to anything open.
//
// Returns closed=true only if an explicit matching </closeName> was consumed.
// closed=false means the caller's tag auto-closes here (EOF or an ancestor's
// close tag, which is left unconsumed for the caller up the stack to match).
func (p *parser) parseContentUntil(closeName string) (children []*Node, closed bool, err error) {
	for {
		tok := p.lex.Peek()
		if tok.EOF() {
			return children, false, nil
		}

		switch tok.Type {
		case p.tt("Text"):
			t := p.lex.Next()
			children = append(children, &Node{Kind: KindText, Text: t.Value})

		case p.tt("TagOpen"):
			checkpoint := p.lex.MakeCheckpoint()
			p.lex.Next() // consume '<'

			next := p.lex.Peek()
			if next.Type == p.tt("Arg") && strings.HasPrefix(next.Value, "/") {
				// closing tag: '<' Arg("/name") (':' arg)* '>'
				nameTok := p.lex.Next()
				name := strings.TrimPrefix(nameTok.Value, "/")

				// Adventure allows (and ignores) arguments echoed back on a
				// close tag, e.g. </color:green> or </italic:false>.
				for p.lex.Peek().Type == p.tt("Colon") {
					p.lex.Next() // consume ':'
					p.lex.Next() // discard the argument token
				}
				p.lex.Next() // consume '>'

				if strings.EqualFold(name, closeName) {
					return children, true, nil
				}
				if closeName == "" {
					// Stray close tag at top level with nothing open:
					// MiniMessage is lenient about invalid tags, so skip it.
					continue
				}
				// Belongs to an ancestor. Rewind so *we* don't consume it, and
				// let the caller (which is closer to matching it) see it.
				p.lex.LoadCheckpoint(checkpoint)
				return children, false, nil
			}

			elem, err := p.parseElement()
			if err != nil {
				return children, false, err
			}
			children = append(children, elem)

		default:
			return children, false, fmt.Errorf("unexpected token %q at %s", tok.Value, tok.Pos)
		}
	}
}

// parseElement parses a tag starting *after* the leading '<' has already been
// consumed by the caller.
func (p *parser) parseElement() (*Node, error) {
	nameTok := p.lex.Next()
	if nameTok.Type != p.tt("Ident") {
		return nil, fmt.Errorf("expected tag name at %s, got %q", nameTok.Pos, nameTok.Value)
	}
	node := &Node{Kind: KindElement, Name: nameTok.Value}

	for p.lex.Peek().Type == p.tt("Colon") {
		p.lex.Next() // consume ':'
		argTok := p.lex.Next()
		switch argTok.Type {
		case p.tt("String"):
			node.Args = append(node.Args, unquote(argTok.Value))
		case p.tt("Arg"), p.tt("Ident"):
			node.Args = append(node.Args, argTok.Value)
		default:
			return nil, fmt.Errorf("expected argument after ':' at %s, got %q", argTok.Pos, argTok.Value)
		}
	}

	closeTok := p.lex.Next()
	switch closeTok.Type {
	case p.tt("SelfClose"):
		node.SelfClosed = true
		node.Closed = true
		return node, nil

	case p.tt("Close"):
		if node.Name == resetTag || leafTags[strings.ToLower(node.Name)] {
			// <reset> and all leafTags (br/newline/lang/tr/translate/...) never
			// have content or a close tag.
			node.SelfClosed = true
			node.Closed = true
			return node, nil
		}
		children, closed, err := p.parseContentUntil(node.Name)
		if err != nil {
			return nil, err
		}
		node.Children = children
		node.Closed = closed
		return node, nil

	default:
		return nil, fmt.Errorf("expected '>' or '/>' at %s, got %q", closeTok.Pos, closeTok.Value)
	}
}

// unquote strips the surrounding quotes from a String token and resolves
// backslash escapes, e.g. "/say \"quoted text\"" -> /say "quoted text".
func unquote(s string) string {
	if len(s) < 2 {
		return s
	}
	inner := s[1 : len(s)-1]
	var b strings.Builder
	for i := 0; i < len(inner); i++ {
		if inner[i] == '\\' && i+1 < len(inner) {
			i++
		}
		b.WriteByte(inner[i])
	}
	return b.String()
}

// --- Demo ----------------------------------------------------------------
