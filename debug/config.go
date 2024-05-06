package debug

import (
	"fmt"
)

type Config struct {
	Port  uint16 `env:"DEBUG_PORT" envDefault:"8080"`
	Debug bool   `env:"DEBUG" envDefault:"false"`
}

func (c *Config) BindAddress() string {
	return fmt.Sprintf("localhost:%d", c.Port)
}
