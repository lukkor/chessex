package chessex

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPGN(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		title          string
		pgn            string
		expected       *PGN
		expectedString string
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
						White:  pString("e4"),
					},
				},
				Outcome: "1/2-1/2",
			},
			`[Event "F/S Return Match"]

1. e4 1/2-1/2`,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			computed := &PGN{}
			err := parser.ParseString("", test.pgn, computed)
			require.NoError(err)
			require.Equal(test.expected, computed)
			require.Equal(test.expectedString, computed.String())
		})
	}
}

func pString(s string) (sr *string) {
	sr = &s
	return sr
}
