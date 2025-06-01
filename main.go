package main

import (
	"strings"

	logger "github.com/sirupsen/logrus"
)

func main() {

	cfg := newConfig()

	switch strings.ToLower(cfg.LogLevel) {
	case "info":
		logger.SetLevel(logger.InfoLevel)
	case "warn":
		logger.SetLevel(logger.WarnLevel)
	case "error":
		logger.SetLevel(logger.ErrorLevel)
	case "fatal":
		logger.SetLevel(logger.FatalLevel)
	case "panic":
		logger.SetLevel(logger.PanicLevel)
	case "trace":
		logger.SetLevel(logger.TraceLevel)
	case "debug":
		logger.SetLevel(logger.DebugLevel)
	default:
		logger.SetLevel(logger.InfoLevel)

	}

	logger.SetFormatter(&logger.JSONFormatter{
		PrettyPrint: true,
	})

	logger.Info()
	if cfg.UnsealKeys == nil {
		logger.Fatal("unseal keys not specified")
	}

	logger.Debug("Vault Unsealer starting...")

	monitorAndUnsealVaults(cfg.Nodes, cfg.UnsealKeys, cfg.ProbeInterval)
}
