package tracing

type Config struct {
	Enabled    bool   `config:"enabled"`
	UseJaeger  bool   `config:"use_jaeger"`
	OutputFile string `config:"output_file"`
}
