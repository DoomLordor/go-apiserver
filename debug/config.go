package debug

import (
	"fmt"
)

type Config struct {
	Active bool   `env:"DEBUG" envDefault:"false"`
	Port   uint16 `env:"DEBUG_PORT" envDefault:"8080"`
}

func (c *Config) BindAddress() string {
	return fmt.Sprintf("localhost:%d", c.Port)
}
