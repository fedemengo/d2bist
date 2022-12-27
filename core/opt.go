package core

type Config struct {
	MaxBits int
}

type Opt func(c *Config)

func WithBitsCap(maxBits int) Opt {
	return func(c *Config) {
		c.MaxBits = maxBits
	}
}
