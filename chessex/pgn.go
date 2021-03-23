// Basics on the PGN format:
//
// The PGN format has 2 parts: tag pairs and movetext.

// Tag pairs are metadata fields enclosed in brackets with the tag name first
// and the value between double quotes.

// Movetext is the full (or partial) representation of the moves with number
// indicators (`1.` for the first move of the game) and the move in Standard
// Algebraic Notation (e.g. `e4`).
//
// See https://en.wikipedia.org/wiki/Portable_Game_Notation for more information.
package chessex

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer/stateful"
)

type PGN struct {
	Tags    []*Tag      `@@*`
	Moves   []*MovePair `@@*`
	Outcome string      `@Outcome`
}

func (pgn *PGN) String() string {
	var s strings.Builder

	for _, tag := range pgn.Tags {
		_, err := s.WriteString(fmt.Sprintf("%s\n", tag))
		if err != nil {
			panic(err)
		}
	}

	_, err := s.WriteString("\n")
	if err != nil {
		panic(err)
	}

	for _, move := range pgn.Moves {
		_, err := s.WriteString(fmt.Sprintf("%s ", move))
		if err != nil {
			panic(err)
		}
	}

	return fmt.Sprintf("%s%s", s.String(), pgn.Outcome)
}

type Tag struct {
	Name  string `"[" @Tag `
	Value string `@String "]"`
}

func (t Tag) String() string {
	return fmt.Sprintf("[%s \"%s\"]", t.Name, t.Value)
}

type MovePair struct {
	Number string `@Number `
	White  *Move  `@@? `
	Black  *Move  `@@?`
}

func (mp *MovePair) String() string {
	return fmt.Sprintf("%s %s %s", mp.Number, mp.White, mp.Black)
}

type Annotation string

const (
	Good        Annotation = "!"
	Excellent              = "!!"
	Mistake                = "?"
	Blunder                = "??"
	Interesting            = "!?"
	Dubious                = "?!"
)

type Move struct {
	Value      string      `( @Move | @Castle | @NullMove )`
	Check      *string     `( @Check )?`
	Annotation *Annotation `( @Annotation )?`
}

func (m *Move) String() string {
	check := ""
	if m.Check != nil {
		check = *m.Check
	}

	annotation := ""
	if m.Annotation != nil {
		annotation = string(*m.Annotation)
	}

	return fmt.Sprintf("%s%s%s", m.Value, check, annotation)
}

var (
	pgnLexer = stateful.MustSimple([]stateful.Rule{
		// Tags
		{"Tag", `Event|Site|Date|Round|Result|WhiteTitle|BlackTitle|WhiteElo|BlackElo|WhiteUSCF|BlackUSCF|WhiteNA|BlackNA|WhiteType|BlackType|EventDate|EventSponsor|Section|Stage|Board|Opening|Variation|SubVariation|ECO|NIC|UTCTime|UTCDate|TimeControl|Time|SetUp|FEN|Termination|Annotator|Mode|PlyCount|Elo|WhiteRatingDiff|BlackRatingDiff|White|Black`, nil},
		{"String", `"(?:\\.|[^"])*"`, nil},
		{"LBracket", `\[`, nil},
		{"RBracket", `\]`, nil},

		// Movetext
		{"Outcome", `(?:1-0|0-1|1/2-1/2|\*)`, nil},
		{"Number", `\d+\.*`, nil},
		{"NullMove", `--`, nil},
		{"Move", `[a-h1-8PNBRQK=x]+`, nil},
		{"Castle", `(?:O-O|O-O-O)`, nil},
		{"Check", `[+#]`, nil},
		{"Annotation", `[?!]+`, nil},
		{"Capture", `x`, nil},
		{"Dot", `\.`, nil},
		{"Slash", `/`, nil},
		{"Dash", `-`, nil},

		{"inlineComment", `{[^}]*}`, nil},
		{"comment", `;[^\n]*\n?`, nil},
		{"whitespace", `[ \n\r]+`, nil},
	})

	parser = participle.MustBuild(
		&PGN{},
		participle.Lexer(pgnLexer),
		participle.Unquote("String"),
		participle.Elide("inlineComment"),
	)
)
