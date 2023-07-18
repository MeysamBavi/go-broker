package metrics

type Config struct {
	Enabled  bool   `config:"enabled"`
	HttpPort string `config:"http_port"`
}
