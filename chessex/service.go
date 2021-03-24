package chessex

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
)

type Service struct {
	cfgPath  string
	printCfg bool
	load     bool

	Cfg    *ServiceCfg
	Log    zerolog.Logger
	Loader *Loader

	Scylla *ScyllaClient

	wg     sync.WaitGroup
	stop   chan struct{}
	status chan error
}

func NewService(cfgPath string, printCfg, load bool) *Service {
	return &Service{
		cfgPath:  cfgPath,
		printCfg: printCfg,
		load:     load,

		stop:   make(chan struct{}),
		status: make(chan error),
	}
}

func (s *Service) Start() error {
	s.Log.Info().Msg("service starting...")

	scylla, err := NewScyllaClient(s)
	if err != nil {
		s.Log.Fatal().Err(err).Msg("cannot create scylla client")
	}

	s.Scylla = scylla

	err = s.Scylla.UpdateSchema()
	if err != nil {
		s.Log.Fatal().Err(err).Msg("cannot update scylla schema")
	}

	s.wg.Add(1)
	go s.loop()

	return nil
}

func (s *Service) Load() error {
	s.Log.Info().Msg("service starting...")

	scylla, err := NewScyllaClient(s)
	if err != nil {
		return fmt.Errorf("cannot create scylla client: %w", err)
	}

	s.Scylla = scylla

	err = s.Scylla.UpdateSchema()
	if err != nil {
		s.Log.Fatal().Err(err).Msg("cannot update scylla schema")
	}

	s.Loader = NewLoader(s)
	if err := s.Loader.Start(); err != nil {
		return fmt.Errorf("cannot load archive: %w", err)
	}

	s.wg.Add(1)
	go s.loop()

	return nil
}

func (s *Service) Stop() {
	close(s.stop)
	s.wg.Wait()

	if s.Loader != nil {
		s.Loader.Stop()
	}

	s.Scylla.Close()

	s.Log.Info().Msg("service stopped")
}

func (s *Service) Die(err error) {
	select {
	case s.status <- err:
		return
	case <-s.stop:
		return
	}
}

func (s *Service) Term() {
	select {
	case s.status <- nil:
		return
	case <-s.stop:
		return
	}
}

func (s *Service) Run() {
	s.Log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	s.Cfg = NewDefaultServiceCfg()

	// Load configuration
	if s.cfgPath != "" {
		err := LoadCfg(s.cfgPath, s.Cfg)
		if err != nil {
			s.Log.Fatal().Err(err).Msg("cannot load config")
		}
	}

	// Print configuration if print-cfg flag is present
	if s.printCfg {
		fmt.Printf(s.Cfg.DumpCfg())
		os.Exit(0)
	}

	if s.load {
		// Load database with archive
		err := s.Load()
		if err != nil {
			s.Log.Fatal().Err(err).Msg("cannot load database")
		}
	} else {
		// Start the service
		err := s.Start()
		if err != nil {
			s.Log.Fatal().Err(err).Msg("cannot start service")
		}
	}

	// Wait for either a signal or an internal error
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-s.status:
		if err != nil {
			s.Log.Error().Err(err).Msg("fatal error")
		}
	case signo := <-sigChan:
		s.Log.Info().Msgf("received signal %d (%v)", signo, signo)
	}

	// Stop the service
	s.Stop()
}

func (s *Service) loop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stop:
			return
		}
	}
}
