package chessex

import (
	"fmt"
	"sync"

	"github.com/alecthomas/participle/v2"
	"github.com/rs/zerolog"
)

type LoaderCfg struct {
	Archive string `json:"archive"`
}

type Loader struct {
	Cfg *LoaderCfg
	Log zerolog.Logger

	Chessex *Service

	parser *participle.Parser
	stop   chan struct{}
	wg     sync.WaitGroup

	games chan *PGN
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

		stop: make(chan struct{}),
	}
}

func (l *Loader) Start() error {
	l.Log.Info().Str("archive", l.Cfg.Archive).Msg("start loading archive...")

	parser, err := NewParser()
	if err != nil {
		return fmt.Errorf("cannot create parser: %w", err)
	}

	l.parser = parser

	if err := l.Load(); err != nil {
		l.Chessex.Die(err)
	}

	return nil
}

func (l *Loader) Load() error {
	l.games = make(chan *PGN, 1)

	go func() {
		archive, err := NewArchive(l.Cfg.Archive)
		if err != nil {
			l.Chessex.Die(err)
		}
		defer archive.Close()

		l.Log.Info().Msg("parser starting...")

		for archive.Scan() {
			pgn := &PGN{}

			err = l.parser.ParseString("", archive.Text(), pgn)
			if err != nil {
				l.Log.Error().Err(err).Msg("cannot parse game")
			}

			l.games <- pgn
		}

		l.games <- nil
	}()

	l.wg.Add(1)
	go l.loop()

	return nil
}

func (l *Loader) Stop() {
	close(l.stop)
	l.wg.Wait()

	l.Log.Info().Msg("loader stopped")
}

func (l *Loader) loop() {
	defer l.wg.Done()

	for {
		select {
		case <-l.stop:
			return

		case game := <-l.games:
			if game == nil {
				l.Chessex.Term()
				return
			}

			l.Log.Info().Str("game", game.String()).Send()
		}
	}
}
