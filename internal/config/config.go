package config

import (
	"fmt"
	"github.com/MeysamBavi/go-broker/api/server"
	"github.com/MeysamBavi/go-broker/internal/store"
	"strings"
)

type Config struct {
	Server server.Config `config:"server"`
	Store  store.Config  `config:"store"`
}

func (c *Config) Validate() error {
	stores := map[string]bool{
		"Store.UseInMemory":  c.Store.UseInMemory,
		"Store.UseCassandra": c.Store.UseCassandra,
	}
	trues := make([]string, 0)
	for s, use := range stores {
		if use {
			trues = append(trues, s)
		}
	}
	if len(trues) == 0 {
		return fmt.Errorf("no store is selected for use")
	}
	if len(trues) > 1 {
		return fmt.Errorf("multiple stores (%s) are selected for use", strings.Join(trues, ", "))
	}

	return nil
}

func Default() Config {
	return Config{
		Server: server.Config{
			Host: "localhost:50043",
		},
		Store: store.Config{
			UseInMemory:  true,
			UseCassandra: false,
			Cassandra: store.CassandraConfig{
				Host:     "localhost:9042",
				Keyspace: "go_broker",
			},
		},
	}
}
