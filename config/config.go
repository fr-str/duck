package config

import "os"

var (
	LogMode  = os.Getenv("LOG_MODE")
	LogLevel = os.Getenv("LOG_LEVEL")
	Port     = os.Getenv("DM_PORT")
)
