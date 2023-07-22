package store

type PostgresConfig struct {
	Host     string `config:"host"`
	Port     string `config:"port"`
	User     string `config:"user"`
	Password string `config:"password"`
	DBName   string `config:"db_name"`
}
