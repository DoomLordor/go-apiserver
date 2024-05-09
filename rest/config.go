package rest

import (
	"fmt"
	"time"
)

type Config struct {
	Active       bool          `env:"REST" envDefault:"false"`
	Host         string        `env:"REST_HOST" envDefault:"localhost"`
	Port         uint16        `env:"REST_PORT" envDefault:"8000"`
	WriteTimeout time.Duration `env:"REST_WRITE_TIMEOUT" envDefault:"15s"`
	ReadTimeout  time.Duration `env:"REST_READ_TIMEOUT" envDefault:"15s"`
	IdleTimeout  time.Duration `env:"REST_IDLE_TIMEOUT" envDefault:"15s"`

	Namespace string `env:"NAMESPACE" envDefault:"test"`
	Subsystem string `env:"SUBSYSTEM" envDefault:"test"`
}

func (c *Config) BindAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
