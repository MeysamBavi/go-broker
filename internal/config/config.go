package config

import (
	"fmt"
	"github.com/MeysamBavi/go-broker/api/server"
	"github.com/MeysamBavi/go-broker/internal/store"
	"github.com/MeysamBavi/go-broker/internal/tracing"
	"github.com/MeysamBavi/go-broker/pkg/metrics"
	"strings"
)

type Config struct {
	Server  server.Config  `config:"server"`
	Store   store.Config   `config:"store"`
	Metrics metrics.Config `config:"metrics"`
	Tracing tracing.Config `config:"tracing"`
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
		Metrics: metrics.Config{
			Enabled:  true,
			HttpPort: "2112",
		},
		Tracing: tracing.Config{
			Enabled:    true,
			UseJaeger:  false,
			OutputFile: "./traces.txt",
		},
	}
}
