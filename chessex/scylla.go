package chessex

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/gocql/gocql"
	"github.com/rs/zerolog"
)

//go:embed data/scylla
var scylla embed.FS

var (
	schemaDir = "data/scylla/schema"
)

type ScyllaCfg struct {
	Hosts []string `json:"hosts"`
}

type ScyllaClient struct {
	Cfg *ScyllaCfg
	Log zerolog.Logger

	session *gocql.Session
}

type sessionFunc func(*gocql.Session)

func NewDefaultScyllaCfg() *ScyllaCfg {
	return &ScyllaCfg{
		Hosts: []string{"127.0.0.1:9042", "127.0.0.1:9043", "127.0.0.1:9044"},
	}
}

// NewScyllaClient creates scylla client
func NewScyllaClient(service *Service) (*ScyllaClient, error) {
	cluster := gocql.NewCluster(service.Cfg.ScyllaCfg.Hosts...)
	cluster.Keyspace = "chessex"
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 30 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("cannot create scylla client session: %w", err)
	}

	return &ScyllaClient{
		Cfg:     service.Cfg.ScyllaCfg,
		Log:     service.Log.With().Str("component", "scylla-client").Logger(),
		session: session,
	}, nil
}

func (sc *ScyllaClient) Close() {
	sc.session.Close()
}

func (sc *ScyllaClient) WithSession(f sessionFunc) {
	f(sc.session)
}

func (sc *ScyllaClient) UpdateSchema() error {
	migrations, err := scylla.ReadDir(schemaDir)
	if err != nil {
		return fmt.Errorf("cannot read schema directory: %w", err)
	}

	for _, migration := range migrations {
		if !migration.IsDir() {
			sc.updateSchema(migration)
		}
	}

	return nil
}

func (sc *ScyllaClient) updateSchema(migration fs.DirEntry) {
	filename := filepath.Join(schemaDir, migration.Name())
	log := sc.Log.With().Str("migration", filename).Logger()

	query, err := scylla.ReadFile(filename)
	if err != nil {
		log.Error().Err(err).Msg("cannot read schema file")
	}

	if err := sc.session.Query(string(query)).Exec(); err != nil {
		log.Error().Err(err).Msg("cannot exec schema query")
	}
}
