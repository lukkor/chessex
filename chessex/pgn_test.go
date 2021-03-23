package chessex

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPGN(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		title    string
		pgn      string
		expected *PGN
	}{
		{
			"Smallest PGN string",
			`[Event "F/S Return Match"]

1. e4 1/2-1/2`,
			&PGN{
				Tags: []*Tag{
					&Tag{
						Name:  "Event",
						Value: "F/S Return Match",
					},
				},
				Moves: []*MovePair{
					&MovePair{
						Number: "1.",
						White: &Move{
							Value:      "e4",
							Check:      nil,
							Annotation: nil,
						},
					},
				},
				Outcome: "1/2-1/2",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			computed := &PGN{}
			err := parser.ParseString("", test.pgn, computed)
			require.NoError(err)
			require.Equal(computed, test.expected)
		})
	}
}
