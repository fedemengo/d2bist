package core

import (
	"github.com/fedemengo/d2bist/pkg/compression"
)

type Config struct {
	InMaxBits         int                         `json:"in_max_bits"`
	InCompressionType compression.CompressionType `json:"in_compression_type"`

	OutMaxBits         int                         `json:"out_max_bits"`
	OutCompressionType compression.CompressionType `json:"out_compression_type"`

	StatsBlockSize    int `json:"stats_block_size"`
	StatsSymbolLen    int `json:"stats_symbol_len"`
	StatsMaxBlockSize int `json:"stats_max_block_size"`
	StatsTopK         int `json:"stats_top_k"`

	EntropyPlotName string `json:"entropy_plot_name"`
}

func NewDefaultConfig() *Config {
	return &Config{
		InMaxBits:         -1,
		InCompressionType: compression.None,

		OutMaxBits:         -1,
		OutCompressionType: compression.None,

		StatsMaxBlockSize: 8,
		StatsTopK:         -1,

		StatsSymbolLen: 2,
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

func WithStatsSymbolLen(entropyChunk int) Opt {
	return func(c *Config) {
		c.StatsSymbolLen = entropyChunk
	}
}

func WithStatsTopK(topK int) Opt {
	return func(c *Config) {
		c.StatsTopK = topK
	}
}

func WithEntropyPlotName(name string) Opt {
	return func(c *Config) {
		c.EntropyPlotName = name
	}
}
