package rest

import (
	"fmt"
	"time"
)

type Config struct {
	Host         string        `env:"HOST" envDefault:"localhost"`
	Port         uint16        `env:"PORT" envDefault:"8000"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"15s"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"15s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" envDefault:"15s"`

	Namespace string `env:"NAMESPACE" envDefault:"test"`
	Subsystem string `env:"SUBSYSTEM" envDefault:"test"`
}

func (c *Config) BindAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
