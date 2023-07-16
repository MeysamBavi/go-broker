package config

import (
	"fmt"
	"github.com/MeysamBavi/go-broker/api/server"
	"github.com/MeysamBavi/go-broker/internal/store"
)

type Config struct {
	Server server.Config `config:"server"`
	Store  store.Config  `config:"store"`
}

func (c *Config) Validate() error {
	if !c.Store.UseInMemory {
		return fmt.Errorf("Store.UseInMemory is set to false, no other store is available")
	}

	return nil
}

func Default() Config {
	return Config{
		Server: server.Config{
			Host: "localhost:50043",
		},
		Store: store.Config{
			UseInMemory: true,
		},
	}
}
