package tracing

type Config struct {
	Enabled         bool   `config:"enabled"`
	UseJaeger       bool   `config:"use_jaeger"`
	OutputFile      string `config:"output_file"`
	JaegerAgentHost string `config:"jaeger_agent_host"`
	JaegerAgentPort string `config:"jaeger_agent_port"`
}
