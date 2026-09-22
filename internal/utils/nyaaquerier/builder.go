package nyaaquerier

import (
	"fmt"
	"strings"
)

// Node is a piece of a nyaa/nekoBT search query.
// The String() output follows the nyaa search syntax:
// https://nyaa.si/help#using-the-search-engine
type Node interface {
	fmt.Stringer
}

// Term is a single word, matched as a substring of the torrent name.
type Term string

func (t Term) String() string { return string(t) }

// Phrase is a quoted multi-word expression, matched as an exact substring.
type Phrase string

func (p Phrase) String() string { return fmt.Sprintf("%q", string(p)) }

// Not negates a node. nyaa only supports negating whole groups,
// so a Not over multiple nodes is wrapped in a group: -"foo bar".
type Not struct{ Node Node }

func (n Not) String() string { return "-" + n.Node.String() }

// And joins nodes with spaces, all of them must match.
type And []Node

func (a And) String() string {
	parts := make([]string, len(a))
	for i, n := range a {
		parts[i] = n.String()
	}
	return strings.Join(parts, " ")
}

// Or joins nodes with |, any of them may match.
// Multi-node alternatives are wrapped in a group so multi-word
// alternatives like (1080 HEVC)|(1080 AV1) keep their words bound together.
// Single nodes are emitted bare: quoted phrases inside parentheses
// lead to unexpected results on nyaa.
type Or []Node

func (o Or) String() string {
	parts := make([]string, len(o))
	for i, n := range o {
		switch n.(type) {
		case And, Group:
			parts[i] = "(" + n.String() + ")"
		default:
			parts[i] = n.String()
		}
	}
	return strings.Join(parts, "|")
}

// Group groups nodes with parentheses for precedence.
// Quoted strings inside parentheses lead to unexpected results on nyaa,
// so phrases are emitted outside of groups.
type Group []Node

func (g Group) String() string {
	parts := make([]string, len(g))
	for i, n := range g {
		parts[i] = n.String()
	}
	return "(" + strings.Join(parts, " ") + ")"
}

// querySanitization strips characters with special meaning in the nyaa
// search syntax. Verified against the live nyaa endpoint: quoted phrases
// containing any of these characters return zero results.
// The dash is intentionally preserved: release names contain dashes
// (e.g. [Erai-raws]) and a dash is only a negation operator when it
// starts a token outside of quotes.
var querySanitization = strings.NewReplacer(
	`"`, " ",
	`'`, " ",
	`(`, " ",
	`)`, " ",
	`:`, " ",
	`,`, " ",
	`/`, " ",
)

// Sanitize strips query-syntax characters from a value and collapses
// the resulting whitespace runs into single spaces.
func Sanitize(value string) string {
	return strings.Join(strings.Fields(querySanitization.Replace(value)), " ")
}

// PhraseOf sanitizes a value and returns it as a Phrase node.
// Multi-word values become quoted phrases, single words become plain terms.
func PhraseOf(value string) Node {
	value = Sanitize(value)
	if strings.ContainsAny(value, " \t") {
		return Phrase(value)
	}
	return Term(value)
}
