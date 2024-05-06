package grpc

import (
	"fmt"
)

type Config struct {
	Host string `env:"HOST" envDefault:"localhost"`
	Port uint16 `env:"PORT" envDefault:"7000"`
}

func (c *Config) BindAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
