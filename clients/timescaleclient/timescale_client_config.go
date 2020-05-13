package timescaleclient

import (
	"fmt"
)

type Config struct {
	User         string
	Password     string
	Host         string
	DatabaseName string
	SLLMode      string
}

func (c *Config) isValid() {
	if c.User == "" || c.Password == "" || c.Host == "" || c.DatabaseName == "" || c.SLLMode == "" {
		panic(fmt.Sprintf("timescale client Config missing: %+v", c))
	}
}
