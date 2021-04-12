package chessex

import (
	"fmt"
	"sync"

	"github.com/alecthomas/participle/v2"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog"
)

type LoaderCfg struct {
	Archive        string `json:"archive"`
	WorkerPoolSize int    `json:"workerPoolSize"`
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
	l.games = make(chan *PGN)

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

			if pgn.Outcome == "*" {
				continue
			}

			l.games <- pgn
		}

		l.Chessex.Term()
	}()

	for range make([]struct{}, l.Cfg.WorkerPoolSize) {
		l.wg.Add(1)
		go l.worker()
	}

	l.wg.Add(1)
	go l.loop()

	return nil
}

func (l *Loader) Stop() {
	close(l.stop)
	l.wg.Wait()

	l.Log.Info().Msg("loader stopped")
}

func (l *Loader) worker() {
	defer l.wg.Done()

	for {
		select {
		case <-l.stop:
			return
		case game := <-l.games:
			l.Chessex.Scylla.WithSession(func(session *gocql.Session) {
				err := game.InsertDepth(session, 3)
				if err != nil {
					l.Log.Error().Err(err).Msg("cannot insert game")
				}
			})
		}
	}
}

func (l *Loader) loop() {
	defer l.wg.Done()

	for {
		select {
		case <-l.stop:
			return
		}
	}
}
