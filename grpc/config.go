package grpc

import (
	"fmt"
)

type Config struct {
	Host string `env:"GRPC_HOST" envDefault:"localhost"`
	Port uint16 `env:"GRPC_PORT" envDefault:"7000"`
}

func (c *Config) BindAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
