package micromessage

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type TokenType int

const (
	TokenText TokenType = iota
	TokenOpenTag
	TokenOpenCloseTag
	TokenCloseTag
	TokenTagValue
)

const (
	tagStart  = '<'
	tagEnd    = '>'
	closeTag  = '/'
	separator = ':'
	escape    = '\\'
)

type token struct {
	start, end int
	typ        TokenType
	children   []*token
}

func (t *token) text(msg []rune) string {
	return string(msg[t.start:t.end])
}

type ParseError struct {
	Message string
}

func (e *ParseError) Error() string { return e.Message }

// Tracks the outer tokenizer state machine.
type firstPassState int

const (
	fpsNormal firstPassState = iota
	fpsTag
	fpsString
)

func tokenize(msg []rune) []*token {
	tokens := firstPass(msg)
	secondPass(msg, tokens)
	return tokens
}

func firstPass(msg []rune) []*token {
	var tokens []*token
	state := fpsNormal
	escaped := false

	currentTokenEnd := 0
	marker := -1
	var currentStringChar rune

	length := len(msg)
	lastEnd := -1

	emit := func(start, end int, typ TokenType) *token {
		t := &token{start: start, end: end, typ: typ}
		tokens = append(tokens, t)
		lastEnd = end
		return t
	}

	for i := 0; i < length; i++ {
		cp := msg[i]

		if !escaped {
			if cp == escape && i+1 < length {
				next := msg[i+1]
				escapeNow := false
				switch state {
				case fpsNormal:
					escapeNow = next == tagStart || next == escape
				case fpsString:
					escapeNow = currentStringChar == next || next == escape
				case fpsTag:
					if next == tagStart {
						escapeNow = true
						state = fpsNormal
					}
				}
				if escapeNow {
					escaped = true
					continue
				}
			}
		} else {
			escaped = false
			continue
		}

		switch state {
		case fpsNormal:
			if cp == tagStart {
				marker = i
				state = fpsTag
			}
		case fpsTag:
			switch cp {
			case tagEnd:
				if i == marker+1 {
					state = fpsNormal
					break
				}
				if currentTokenEnd != marker {
					emit(currentTokenEnd, marker, TokenText)
				}
				currentTokenEnd = i + 1

				typ := TokenOpenTag
				if marker+1 < length && msg[marker+1] == closeTag {
					typ = TokenCloseTag
				} else if i-1 >= 0 && msg[i-1] == closeTag {
					typ = TokenOpenCloseTag
				}
				emit(marker, currentTokenEnd, typ)
				state = fpsNormal
			case tagStart:
				marker = i
			case '\'', '"':
				currentStringChar = cp
				if strings.ContainsRune(string(msg[i+1:]), cp) {
					state = fpsString
				}
			}
		case fpsString:
			if cp == currentStringChar {
				state = fpsTag
			}
		}

		if i == length-1 && state == fpsTag {
			i = marker
			state = fpsNormal
		}
	}

	if lastEnd == -1 {
		emit(0, length, TokenText)
	} else if lastEnd != length {
		emit(lastEnd, length, TokenText)
	}

	return tokens
}

type secondPassState int

const (
	spsNormal secondPassState = iota
	spsString
)

func secondPass(msg []rune, tokens []*token) {
	for _, tok := range tokens {
		if tok.typ != TokenOpenTag && tok.typ != TokenOpenCloseTag && tok.typ != TokenCloseTag {
			continue
		}

		startIndex := tok.start + 1
		if tok.typ == TokenCloseTag {
			startIndex = tok.start + 2
		}
		endIndex := tok.end - 1
		if tok.typ == TokenOpenCloseTag {
			endIndex = tok.end - 2
		}

		state := spsNormal
		escaped := false
		var currentStringChar rune
		marker := startIndex

		insert := func(t *token) {
			tok.children = append(tok.children, t)
		}

		for i := startIndex; i < endIndex; i++ {
			cp := msg[i]

			if !escaped {
				if cp == escape && i+1 < len(msg) {
					next := msg[i+1]
					escapeNow := false
					switch state {
					case spsNormal:
						escapeNow = next == tagStart || next == escape
					case spsString:
						escapeNow = currentStringChar == next || next == escape
					}
					if escapeNow {
						escaped = true
						continue
					}
				}
			} else {
				escaped = false
				continue
			}

			switch state {
			case spsNormal:
				if cp == separator {
					if i+2 < len(msg) && msg[i+1] == '/' && msg[i+2] == '/' {
						continue
					}
					if marker == i {
						insert(&token{start: i, end: i, typ: TokenTagValue})
						marker++
					} else {
						insert(&token{start: marker, end: i, typ: TokenTagValue})
						marker = i + 1
					}
				} else if cp == '\'' || cp == '"' {
					state = spsString
					currentStringChar = cp
				}
			case spsString:
				if cp == currentStringChar {
					state = spsNormal
				}
			}
		}

		if len(tok.children) == 0 {
			insert(&token{start: startIndex, end: endIndex, typ: TokenTagValue})
		} else {
			last := tok.children[len(tok.children)-1]
			if last.end != endIndex {
				insert(&token{start: last.end + 1, end: endIndex, typ: TokenTagValue})
			}
		}
	}
}

// Strips a single layer of matching quotes from a token and resolves backslash escapes inside it.
func unquoteAndEscape(msg []rune, start, end int) string {
	if end > start && (msg[start] == '\'' || msg[start] == '"') && end-1 > start && msg[end-1] == msg[start] {
		start++
		end--
	}
	var sb strings.Builder
	for i := start; i < end; i++ {
		if msg[i] == escape && i+1 < end {
			sb.WriteRune(msg[i+1])
			i++
			continue
		}
		sb.WriteRune(msg[i])
	}
	return sb.String()
}

// A node in the intermediate parse tree
type elementNode struct {
	parent   *elementNode
	children []*elementNode

	// text nodes
	isText bool
	text   string

	// tag nodes
	tagName   string
	tagArgs   []string
	selfClose bool
}

func (n *elementNode) addChild(c *elementNode) {
	c.parent = n
	n.children = append(n.children, c)
}

// Creates the element tree from a MiniMessage string
func parseTree(input string, strict bool) (*elementNode, error) {
	msg := []rune(input)
	tokens := tokenize(msg)

	root := &elementNode{isText: false, tagName: ""}
	node := root

	for _, tok := range tokens {
		switch tok.typ {
		case TokenText:
			node.addChild(&elementNode{isText: true, text: unescapeText(msg, tok.start, tok.end)})

		case TokenOpenTag, TokenOpenCloseTag:
			if len(tok.children) == 0 {
				node.addChild(&elementNode{isText: true, text: tok.text(msg)})
				continue
			}
			nameTok := tok.children[0]
			name := unquoteAndEscape(msg, nameTok.start, nameTok.end)
			if !isValidTagName(name) {
				node.addChild(&elementNode{isText: true, text: tok.text(msg)})
				continue
			}

			if strings.EqualFold(name, "reset") {
				if strict {
					return nil, &ParseError{Message: "<reset> tags are not allowed when strict mode is enabled"}
				}
				node = root
				continue
			}

			var args []string
			for _, c := range tok.children[1:] {
				args = append(args, unquoteAndEscape(msg, c.start, c.end))
			}

			tagNode := &elementNode{tagName: name, tagArgs: args, selfClose: tok.typ == TokenOpenCloseTag}
			node.addChild(tagNode)
			if tok.typ != TokenOpenCloseTag {
				node = tagNode
			}

		case TokenCloseTag:
			if len(tok.children) == 0 {
				node.addChild(&elementNode{isText: true, text: tok.text(msg)})
				continue
			}
			nameTok := tok.children[0]
			closeName := unquoteAndEscape(msg, nameTok.start, nameTok.end)

			if strings.EqualFold(closeName, "reset") {
				continue
			}

			// Walk up looking for a matching open tag
			found := false
			p := node
			for p != root && p != nil {
				if strings.EqualFold(p.tagName, closeName) {
					if strict && p != node {
						return nil, &ParseError{Message: fmt.Sprintf("Unclosed tag encountered; %s is not closed, because %s was closed first.", node.tagName, closeName)}
					}
					node = p.parent
					found = true
					break
				}
				p = p.parent
			}
			if !found {
				node.addChild(&elementNode{isText: true, text: tok.text(msg)})
			}
		}
	}

	if strict && node != root {
		return nil, &ParseError{Message: fmt.Sprintf("Unclosed tag: %s", node.tagName)}
	}

	return root, nil
}

// Resolves backslash escapes within a plain-text token.
func unescapeText(msg []rune, start, end int) string {
	var sb strings.Builder
	for i := start; i < end; i++ {
		if msg[i] == escape && i+1 < end {
			next := msg[i+1]
			if next == tagStart || next == escape {
				sb.WriteRune(next)
				i++
				continue
			}
		}
		sb.WriteRune(msg[i])
	}
	return sb.String()
}

func isValidTagName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r == ' ' || r == tagStart || r == tagEnd {
			return false
		}
		if i == 0 && (r == '#' || r == '!') {
			continue
		}
	}
	return true
}

func argAsFloat64(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func codePointLen(s string) int {
	return utf8.RuneCountInString(s)
}
