package store

type Config struct {
	UseInMemory  bool            `config:"in_memory"`
	UseCassandra bool            `config:"use_cassandra"`
	Cassandra    CassandraConfig `config:"cassandra"`
}
