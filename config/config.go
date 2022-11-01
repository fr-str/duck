package config

import (
	"github.com/timoni-io/go-utils/env"
)

var (
	LogMode  = env.Get("LOG_MODE", "prod")
	LogLevel = env.Get("LOG_LEVEL", "i")
	Port     = func() string {
		p := env.Get("DP_PORT", "6666")
		if p[0] != ':' {
			p = ":" + p
		}
		return p
	}()
)
