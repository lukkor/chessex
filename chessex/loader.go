package chessex

import (
	"bufio"
	"compress/bzip2"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/rs/zerolog"
)

type LoaderCfg struct {
	Archive string `json:"archive"`
}

type Loader struct {
	Cfg *LoaderCfg
	Log zerolog.Logger

	Chessex *Service
}

func NewDefaultLoaderCfg() *LoaderCfg {
	return &LoaderCfg{
		Archive: "",
	}
}

func NewLoader(service *Service) *Loader {
	return &Loader{
		Cfg: service.Cfg.LoaderCfg,
		Log: service.Log.With().Str("component", "loader").Logger(),

		Chessex: service,
	}
}

func (l *Loader) Start() error {
	l.Log.Info().Str("archive", l.Cfg.Archive).Msg("start loading archive...")

	go func() {
		if err := l.Load(); err != nil {
			l.Chessex.Die(err)
		}

		t := time.NewTimer(2 * time.Second)
		defer t.Stop()

		select {
		case <-t.C:
			l.Chessex.Term()
		}
	}()

	return nil
}

func (l *Loader) Load() error {
	// Open the archive
	file, err := os.Open(l.Cfg.Archive)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create an unzip stream reader
	r := bzip2.NewReader(file)

	// Scan for each game in the stream reader
	scanner := newGameScanner(r)
	i := 0
	for scanner.Scan() {
		pgn := &PGN{}
		text := scanner.Text()
		err := parser.ParseString("", text, pgn)
		if err != nil {
			return fmt.Errorf("%s %w", text, err)
		}

		l.Log.Info().Interface("game", pgn.String()).Send()

		if i == 5 {
			break
		}

		i++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func newGameScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(splitGame)

	return scanner
}

func splitGame(data []byte, atEOF bool) (advance int, token []byte, err error) {
	endOfGame := regexp.MustCompile(`[\r\n]{2}\[`)

	if len(data) == 0 {
		return 0, nil, nil
	}

	if loc := endOfGame.FindIndex(data); loc != nil && loc[0] >= 0 {
		return loc[1], data[0:loc[0]], nil
	}

	if atEOF {
		return len(data), data, io.EOF
	}

	return 0, nil, nil
}
