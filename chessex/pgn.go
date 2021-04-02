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
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer/stateful"
	"github.com/gocql/gocql"
)

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
		{"Move", `[a-h1-8PNBRQK=x+#!?]+`, nil},
		{"Castle", `(?:O-O-O|O-O)[+#!?]*`, nil},
		{"NullMove", `--`, nil},

		{"inlineComment", `{[^}]*}`, nil},
		{"comment", `;[^\n]*\n?`, nil},
		{"whitespace", `[ \n\r]+`, nil},
	})
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
	Number string  `@Number `
	White  *string `( @Move | @Castle | @NullMove )?`
	Black  *string `( @Move | @Castle | @NullMove )?`
}

func (mp *MovePair) String() string {
	white := ""
	if mp.White != nil {
		white = fmt.Sprintf(" %s", *mp.White)
	}

	black := ""
	if mp.Black != nil {
		black = fmt.Sprintf(" %s", *mp.Black)
	}

	return fmt.Sprintf("%s%s%s", mp.Number, white, black)
}

func NewParser() (*participle.Parser, error) {
	parser, err := participle.Build(
		&PGN{},
		participle.Lexer(pgnLexer),
		participle.Unquote("String"),
		participle.Elide("inlineComment"),
	)
	if err != nil {
		return nil, err
	}

	return parser, nil
}

func (pgn *PGN) Insert(session *gocql.Session) error {
	query := `INSERT INTO games_by_opening (id, opening, outcome, tags, raw) VALUES (?, ?, ?, ?, ?)`

	if len(pgn.Moves) == 0 {
		return fmt.Errorf("cannot insert game without opening (0 move)")
	}

	if pgn.Moves[0].White == nil {
		return fmt.Errorf("cannot insert game without opening (white nil)")
	}

	opening := pgn.Moves[0].White

	tags := map[string]string{}
	for _, tag := range pgn.Tags {
		tags[tag.Name] = tag.Value
	}

	id := sha256.Sum256([]byte(pgn.String()))

	return session.Query(query, id[:], opening, pgn.Outcome, tags, pgn.String()).Exec()
}
