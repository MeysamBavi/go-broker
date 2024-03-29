package store

import "github.com/MeysamBavi/go-broker/internal/store/batch"

type Config struct {
	UseInMemory  bool            `config:"in_memory"`
	UseCassandra bool            `config:"use_cassandra"`
	Cassandra    CassandraConfig `config:"cassandra"`
	UsePostgres  bool            `config:"use_postgres"`
	Postgres     PostgresConfig  `config:"postgres"`
	Batch        batch.Config    `config:"batch"`
}
