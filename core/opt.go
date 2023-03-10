package core

import "github.com/fedemengo/d2bist/compression"

type Config struct {
	InMaxBits         int
	InCompressionType compression.CompressionType

	OutMaxBits         int
	OutCompressionType compression.CompressionType
}

func NewDefaultConfig() *Config {
	return &Config{
		InMaxBits:         -1,
		InCompressionType: compression.None,

		OutMaxBits:         -1,
		OutCompressionType: compression.None,
	}
}

type Opt func(c *Config)

func WithOutBitsCap(maxBits int) Opt {
	return func(c *Config) {
		c.OutMaxBits = maxBits
	}
}

func WithOutCompression(ct compression.CompressionType) Opt {
	return func(c *Config) {
		c.OutCompressionType = ct
	}
}

func WithInBitsCap(maxBits int) Opt {
	return func(c *Config) {
		c.InMaxBits = maxBits
	}
}

func WithInCompression(ct compression.CompressionType) Opt {
	return func(c *Config) {
		c.InCompressionType = ct
	}
}
