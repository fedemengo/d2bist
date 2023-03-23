package core

import "github.com/fedemengo/d2bist/compression"

type Config struct {
	InMaxBits         int
	InCompressionType compression.CompressionType

	OutMaxBits         int
	OutCompressionType compression.CompressionType

	StatsBlockSize    int
	StatsEntropyChunk int
	StatsMaxBlockSize int
	StatsTopK         int
}

func NewDefaultConfig() *Config {
	return &Config{
		InMaxBits:         -1,
		InCompressionType: compression.None,

		OutMaxBits:         -1,
		OutCompressionType: compression.None,

		StatsMaxBlockSize: 8,
		StatsTopK:         -1,
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

func WithStatsMaxBlockSize(maxBlockSize int) Opt {
	return func(c *Config) {
		c.StatsMaxBlockSize = maxBlockSize
	}
}

func WithStatsBlockSize(blockSize int) Opt {
	return func(c *Config) {
		c.StatsBlockSize = blockSize
	}
}

func WithStatsEntropyChunk(entropyChunk int) Opt {
	return func(c *Config) {
		c.StatsEntropyChunk = entropyChunk
	}
}

func WithStatsTopK(topK int) Opt {
	return func(c *Config) {
		c.StatsTopK = topK
	}
}
