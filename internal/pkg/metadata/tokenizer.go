package metadata

import (
	"strings"
	"unicode/utf8"
)

type TokenKind uint8

const (
	TokenText TokenKind = iota
	TokenBracket
	TokenParen
	TokenPipe
	TokenTagBlock
)

type Token struct {
	Kind        TokenKind
	Text        string
	Open, Close rune
}

// Tokenize performs lossless, nesting-aware tokenization. It deliberately does not
// assign semantics; Parse does that in a later pass.
func Tokenize(s string) []Token {
	var out []Token
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		kind := TokenText
		var close rune
		switch r {
		case '[':
			kind, close = TokenBracket, ']'
		case '(':
			kind, close = TokenParen, ')'
		case '{':
			kind, close = TokenTagBlock, '}'
		case '|':
			kind = TokenPipe
		}
		if kind == TokenText {
			j := i + n
			for j < len(s) {
				rr, nn := utf8.DecodeRuneInString(s[j:])
				if strings.ContainsRune("[({|", rr) {
					break
				}
				j += nn
			}
			out = append(out, Token{Kind: TokenText, Text: s[i:j]})
			i = j
			continue
		}
		if kind == TokenPipe {
			out = append(out, Token{Kind: kind, Text: "|", Open: r})
			i += n
			continue
		}
		j := matchingClose(s, i, r, close)
		if j >= 0 {
			out = append(out, Token{Kind: kind, Text: s[i+n : j], Open: r, Close: close})
			i = j + n
			continue
		}
		out = append(out, Token{Kind: kind, Text: string(r), Open: r})
		i += n
	}
	return out
}
