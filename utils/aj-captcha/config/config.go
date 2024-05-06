package config

import (
	"errors"
	"issue_pr_board/utils/aj-captcha/const"
	"strings"
)

type BlockPuzzleConfig struct {
	Offset int `yaml:"offset"`
}

type Config struct {
	BlockPuzzle    *BlockPuzzleConfig `yaml:"blockPuzzle"`
	CacheType      string             `yaml:"cacheType"`
	CacheExpireSec int                `yaml:"cacheExpireSec"`
	ResourcePath   string             `yaml:"resourcePath"`
}

func BuildConfig(cacheType, resourcePath string, puzzleConfig *BlockPuzzleConfig, cacheExpireSec int) *Config {
	if len(resourcePath) == 0 {
		resourcePath = constant.DefaultResourceRoot
	}
	if len(cacheType) == 0 {
		cacheType = constant.MemCacheKey
	} else if strings.Compare(cacheType, constant.MemCacheKey) != 0 {
		panic(errors.New("cache type not support"))
		return nil
	}
	if cacheExpireSec == 0 {
		cacheExpireSec = 2 * 60
	}
	if nil == puzzleConfig {
		puzzleConfig = &BlockPuzzleConfig{Offset: 10}
	}

	return &Config{
		CacheType:      cacheType,
		BlockPuzzle:    puzzleConfig,
		CacheExpireSec: cacheExpireSec,
		ResourcePath:   resourcePath,
	}
}
