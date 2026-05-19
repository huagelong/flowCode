package bootstrap

import "time"

type Config struct {
	HTTP HTTPConfig
	DB   DBConfig
}

type HTTPConfig struct {
	Addr string
}

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func LoadConfig() (*Config, error) {
	return &Config{
		HTTP: HTTPConfig{Addr: ":8080"},
		DB: DBConfig{
			DSN:             "root:password@tcp(127.0.0.1:3306)/anserflow?charset=utf8mb4&parseTime=True&loc=Local",
			MaxOpenConns:    20,
			MaxIdleConns:    10,
			ConnMaxLifetime: time.Hour,
		},
	}, nil
}
