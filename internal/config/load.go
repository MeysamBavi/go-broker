package config

import (
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"log"
	"strings"
)

const (
	tag                = "config"
	delimiter          = "."
	prefix             = "GO_BROKER__"
	separator          = "__"
	jsonConfigFilePath = "config.json"
)

func Load() *Config {
	k := koanf.New(delimiter)

	{
		err := k.Load(structs.Provider(Default(), tag), nil)
		if err != nil {
			log.Fatalf("could not load default config: %v\n", err)
		}
	}

	{
		err := k.Load(file.Provider(jsonConfigFilePath), json.Parser())
		if err != nil {
			log.Printf("could not load json config: %v\n", err)
		}
	}

	{
		err := k.Load(env.Provider(prefix, delimiter, envCallBack), nil)
		if err != nil {
			log.Printf("could not load env variables for config: %v\n", err)
		}
	}

	var instance Config
	err := k.UnmarshalWithConf("", &instance, koanf.UnmarshalConf{
		Tag: tag,
	})

	if err != nil {
		log.Fatalf("could not unmarshal config: %v\n", err)
	}

	if err := instance.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v\n", err)
	}

	return &instance
}

func envCallBack(s string) string {
	base := strings.ToLower(strings.TrimPrefix(s, prefix))

	return strings.ReplaceAll(base, separator, delimiter)
}
