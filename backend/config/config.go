package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct{ Port int }

func Load() Config {
	p := 8080
	if v, e := strconv.Atoi(os.Getenv("PORT")); e == nil && v > 0 && v < 65536 {
		p = v
	}
	return Config{p}
}
func (c Config) Address() string { return fmt.Sprintf(":%d", c.Port) }
