package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"log"
	"time"
)

const (
	tag                = "client"
	delimiter          = "."
	yamlConfigFilePath = "config.yaml"
)

type Config struct {
	Connections     int           `client:"connections"`
	StartRPS        float64       `client:"start_rps"`
	TargetRPS       float64       `client:"target_rps"`
	RiseDuration    time.Duration `client:"rise_duration"`
	PlateauDuration time.Duration `client:"plateau_duration"`
}

func Default() Config {
	return Config{
		Connections:     10,
		StartRPS:        2,
		TargetRPS:       4,
		RiseDuration:    30 * time.Second,
		PlateauDuration: 5 * time.Second,
	}
}

func Load() *Config {
	k := koanf.New(delimiter)

	{
		err := k.Load(structs.Provider(Default(), tag), nil)
		if err != nil {
			log.Fatalf("could not load default config: %v\n", err)
		}
	}

	{
		err := k.Load(file.Provider(yamlConfigFilePath), yaml.Parser())
		if err != nil {
			log.Fatalf("could not load yaml config: %v\n", err)
		}
	}

	var instance Config
	err := k.UnmarshalWithConf("", &instance, koanf.UnmarshalConf{
		Tag: tag,
	})

	if err != nil {
		log.Fatalf("could not unmarshal config: %v\n", err)
	}

	return &instance
}
