package micromessage

import (
	"fmt"
	"strings"
)

// This is a direct port of Adventure's real MiniMessage tokenizer/tree
// builder (TokenParser.parseString / TokenParser.buildTree), rather than a
// generic recursive-descent parser, specifically to match its lenient,
// single-pass behavior: tag resolution happens *during* parsing (so a tag
// that doesn't resolve to anything becomes literal text using its exact
// source span, and self-closing-ness is just a property of the resolved
// Tag), a stray or mismatched close tag never errors (it's literal text, or
// closes whichever open ancestor it actually matches), and a ParserDirective
// like <reset> doesn't appear in the tree at all -- it collapses every
// currently open tag back to the root right where it occurs, which is what
// lets <reset> correctly terminate an enclosing <gradient>/<rainbow>.

// --- AST ---------------------------------------------------------------

type Kind int

const (
	KindText Kind = iota
	KindElement
)

// Node is either a run of text (which may be the exact source text of a tag
// that didn't resolve to anything) or a resolved tag.
type Node struct {
	Kind Kind

	Text string // set when Kind == KindText

	Name     string // tag name, set when Kind == KindElement
	Args     []string
	Tag      Tag // the Tag this element resolved to, already looked up at parse time
	Children []*Node
}

// --- Tokenizer -----------------------------------------------------------

type tokenKind int

const (
	tokText tokenKind = iota
	tokOpenTag
	tokCloseTag
	tokOpenCloseTag
)

// token spans [start,end) in the source rune slice; for tag tokens that
// includes the surrounding '<'/'>'/'/' characters.
type token struct {
	kind       tokenKind
	start, end int
}

// tokenize is a direct port of TokenParser.parseString: a single forward
// scan that finds '<'...'>' spans, treating a quote character as opening a
// nested "string" region (so '<'/'>' inside it don't end the tag) only if a
// matching closing quote actually appears later in the same tag, and
// honoring '\' to escape '<' and itself in plain text, or the active quote
// character and '\' inside a string.
func tokenize(src []rune) []token {
	const (
		stNormal = iota
		stTag
		stString
	)

	var tokens []token
	state := stNormal
	escaped := false
	currentTokenEnd := 0
	marker := -1
	var quoteChar rune

	n := len(src)
	for i := 0; i < n; i++ {
		r := src[i]

		if !escaped {
			if r == '\\' && i+1 < n {
				next := src[i+1]
				switch state {
				case stNormal:
					escaped = next == '<' || next == '\\'
				case stString:
					escaped = next == quoteChar || next == '\\'
				case stTag:
					if next == '<' {
						escaped = true
						state = stNormal
					}
				}
				if escaped {
					continue
				}
			}
		} else {
			escaped = false
			continue
		}

		switch state {
		case stNormal:
			if r == '<' {
				marker = i
				state = stTag
			}

		case stTag:
			switch r {
			case '>':
				if i == marker+1 {
					// "<>" is empty, not a tag.
					state = stNormal
					break
				}
				if currentTokenEnd != marker {
					tokens = append(tokens, token{tokText, currentTokenEnd, marker})
				}
				end := i + 1
				currentTokenEnd = end

				kind := tokOpenTag
				if marker+1 < n && src[marker+1] == '/' {
					kind = tokCloseTag
				} else if i-1 >= 0 && src[i-1] == '/' {
					kind = tokOpenCloseTag
				}
				tokens = append(tokens, token{kind, marker, end})
				state = stNormal

			case '<':
				// Not a tag after all, but a new one could start here.
				marker = i

			case '\'', '"':
				quoteChar = r
				if runeIndexFrom(src, r, i+1) != -1 {
					state = stString
				}
			}

		case stString:
			if r == quoteChar {
				state = stTag
			}
		}

		if i == n-1 && state == stTag {
			// Reached EOF with an open '<' never matched to a '>'. Rewind to
			// just after it, back in NORMAL state, so anything that looked
			// like it was inside a quoted string (and so was skipped above)
			// gets a chance to be scanned for tags too.
			i = marker
			state = stNormal
		}
	}

	lastEnd := 0
	if len(tokens) > 0 {
		lastEnd = tokens[len(tokens)-1].end
	}
	if lastEnd != n {
		tokens = append(tokens, token{tokText, lastEnd, n})
	}
	return tokens
}

func runeIndexFrom(src []rune, r rune, from int) int {
	for i := from; i < len(src); i++ {
		if src[i] == r {
			return i
		}
	}
	return -1
}

// tagInner returns the [start,end) span of a tag token's content, excluding
// the surrounding '<', '>', and '/' characters.
func tagInner(tok token) (start, end int) {
	start = tok.start + 1
	if tok.kind == tokCloseTag {
		start = tok.start + 2
	}
	end = tok.end - 1
	if tok.kind == tokOpenCloseTag {
		end = tok.end - 2
	}
	return start, end
}

// splitTagParts is a direct port of TokenParser's second pass: splits a
// tag's inner content on ':', except a ':' immediately followed by "//"
// (part of a URL scheme, e.g. <click:open_url:https://example.com> needs no
// quoting), and except inside a quoted string. Always returns at least one
// part (the whole span, if no ':' was found).
func splitTagParts(src []rune, start, end int) [][2]int {
	const (
		stNormal = iota
		stString
	)
	state := stNormal
	escaped := false
	var quoteChar rune
	marker := start
	var parts [][2]int

	for i := start; i < end; i++ {
		r := src[i]

		if !escaped {
			if r == '\\' && i+1 < len(src) {
				next := src[i+1]
				switch state {
				case stNormal:
					escaped = next == '<' || next == '\\'
				case stString:
					escaped = next == quoteChar || next == '\\'
				}
				if escaped {
					continue
				}
			}
		} else {
			escaped = false
			continue
		}

		switch state {
		case stNormal:
			switch r {
			case ':':
				if i+2 < len(src) && src[i+1] == '/' && src[i+2] == '/' {
					continue
				}
				parts = append(parts, [2]int{marker, i})
				marker = i + 1
			case '\'', '"':
				state = stString
				quoteChar = r
			}
		case stString:
			if r == quoteChar {
				state = stNormal
			}
		}
	}
	parts = append(parts, [2]int{marker, end})
	return parts
}

// unescapeText strips a '\' that precedes '<' or another '\', matching
// Adventure's TextNode (every plain-text span, including a tag's raw source
// text when it's rendered literally because it didn't resolve to anything,
// goes through this same unescape pass).
func unescapeText(src []rune, start, end int) string {
	var b strings.Builder
	for i := start; i < end; i++ {
		r := src[i]
		if r == '\\' && i+1 < end && (src[i+1] == '<' || src[i+1] == '\\') {
			i++
			b.WriteRune(src[i])
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// unquoteAndEscape strips a leading/trailing matching quote character (if
// any) and, only within a quoted span, unescapes '\' followed by the quote
// character or another '\'. An unquoted (bare) argument is returned exactly
// as written -- it is not escape-processed at all, matching Adventure.
func unquoteAndEscape(src []rune, start, end int) string {
	if start == end {
		return ""
	}
	startIndex, endIndex := start, end
	first := src[startIndex]
	last := src[endIndex-1]
	if first == '\'' || first == '"' {
		startIndex++
	} else {
		return string(src[startIndex:endIndex])
	}
	if last == '\'' || last == '"' {
		endIndex--
	}
	if startIndex > endIndex {
		return string(src[start:end])
	}
	var b strings.Builder
	for i := startIndex; i < endIndex; i++ {
		r := src[i]
		if r == '\\' && i+1 < endIndex && (src[i+1] == first || src[i+1] == '\\') {
			i++
			b.WriteRune(src[i])
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// validTagName matches Adventure's tag-name pattern: an optional leading
// '!'/'?'/'#' sigil followed by any number of lowercase letters, digits,
// '_', or '-' (checked case-insensitively).
func validTagName(name string) bool {
	i := 0
	if i < len(name) && strings.ContainsRune("!?#", rune(name[i])) {
		i++
	}
	for ; i < len(name); i++ {
		c := name[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if !(c == '_' || c == '-' || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// --- Tree builder ----------------------------------------------------------

// openFrame is an ancestor tag still waiting for a matching close (or EOF).
type openFrame struct {
	node *Node
	name string // lower-cased, for case-insensitive close matching
	args []string
}

// parse tokenizes and builds src into a tree of top-level nodes, resolving
// every tag against resolvers as it goes (so an unrecognized/invalid tag
// becomes a literal-text Node using its exact source span, never a parse
// error) matching Adventure's default (non-strict) leniency. In strict
// mode, a <reset> tag or an unclosed tag at EOF is a hard error instead.
func parse(src string, resolvers []TagResolver, strict bool) ([]*Node, error) {
	runes := []rune(src)
	tokens := tokenize(runes)

	root := &Node{Kind: KindElement}
	cursor := root
	var stack []*openFrame

	literal := func(tok token) *Node {
		return &Node{Kind: KindText, Text: unescapeText(runes, tok.start, tok.end)}
	}
	appendChild := func(n *Node) {
		cursor.Children = append(cursor.Children, n)
	}
	partsToArgs := func(parts [][2]int) []string {
		args := make([]string, len(parts))
		for i, p := range parts {
			args[i] = unquoteAndEscape(runes, p[0], p[1])
		}
		return args
	}

	for _, tok := range tokens {
		switch tok.kind {
		case tokText:
			appendChild(&Node{Kind: KindText, Text: unescapeText(runes, tok.start, tok.end)})

		case tokOpenTag, tokOpenCloseTag:
			innerStart, innerEnd := tagInner(tok)
			all := partsToArgs(splitTagParts(runes, innerStart, innerEnd))
			name, args := all[0], all[1:]

			if !validTagName(name) {
				appendChild(literal(tok))
				continue
			}
			tag, ok := resolveTag(resolvers, name, args)
			if !ok {
				appendChild(literal(tok))
				continue
			}
			if tag.kind == tagDirective {
				if strict {
					return nil, fmt.Errorf("<%s> tags are not allowed when strict mode is enabled", name)
				}
				cursor = root
				stack = stack[:0]
				continue
			}

			node := &Node{Kind: KindElement, Name: name, Args: args, Tag: tag}
			appendChild(node)
			if tok.kind != tokOpenCloseTag && !tag.selfClosing {
				stack = append(stack, &openFrame{node: node, name: strings.ToLower(name), args: args})
				cursor = node
			}

		case tokCloseTag:
			innerStart, innerEnd := tagInner(tok)
			closeVals := partsToArgs(splitTagParts(runes, innerStart, innerEnd))
			closeName := closeVals[0]

			if !validTagName(closeName) {
				appendChild(literal(tok))
				continue
			}

			// Closing only ever needs to compare names/echoed args against
			// what's actually open -- never call a resolver with the close
			// tag's own (often absent) arguments for this, since a resolver
			// that uses ArgumentQueue.PopOr on a required argument would
			// otherwise panic on a merely-probed empty queue.
			matchIdx := -1
			for i := len(stack) - 1; i >= 0; i-- {
				if tagCloses(closeVals, stack[i].name, stack[i].args) {
					matchIdx = i
					break
				}
			}
			if matchIdx != -1 {
				if strict && matchIdx != len(stack)-1 {
					return nil, fmt.Errorf("unclosed tag <%s>: <%s> was closed first", stack[len(stack)-1].name, closeName)
				}
				stack = stack[:matchIdx]
				if len(stack) == 0 {
					cursor = root
				} else {
					cursor = stack[len(stack)-1].node
				}
				continue
			}

			// Nothing open matches. A stray "</reset>" (or a custom
			// directive spelling) is still a no-op, same as when it's
			// actually closing something -- everything else is literal text.
			if directiveCloseTag(resolvers, closeName, closeVals[1:]) {
				continue
			}
			appendChild(literal(tok))
		}
	}

	if strict && len(stack) > 0 {
		names := make([]string, len(stack))
		for i, f := range stack {
			names[i] = f.name
		}
		return nil, fmt.Errorf("all tags must be explicitly closed while in strict mode, still open: %s", strings.Join(names, ", "))
	}

	return root.Children, nil
}

// tagCloses reports whether a close tag's parts (name plus any echoed
// arguments) close an open tag with the given name/args: the name matches
// case-insensitively, and any echoed close-tag arguments must equal the
// open tag's corresponding arguments exactly.
func tagCloses(closeVals []string, openName string, openArgs []string) bool {
	if len(closeVals) > 1+len(openArgs) {
		return false
	}
	if !strings.EqualFold(closeVals[0], openName) {
		return false
	}
	for i := 1; i < len(closeVals); i++ {
		if closeVals[i] != openArgs[i-1] {
			return false
		}
	}
	return true
}

// directiveCloseTag reports whether name/args resolve to a ParserDirective,
// for recognizing a stray "</reset>"-like close tag as a no-op even when
// nothing open matches it. A resolver that panics while being probed this
// way (e.g. ArgumentQueue.PopOr on a required argument the close tag didn't
// echo) is treated as "not a directive" rather than aborting the parse --
// this is only a best-effort convenience lookup, not a real resolution.
func directiveCloseTag(resolvers []TagResolver, name string, args []string) (isDirective bool) {
	defer func() {
		if recover() != nil {
			isDirective = false
		}
	}()
	tag, ok := resolveTag(resolvers, name, args)
	return ok && tag.kind == tagDirective
}
