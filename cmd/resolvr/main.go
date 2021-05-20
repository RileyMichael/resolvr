package main

import (
	"github.com/RileyMichael/resolvr/internal/resolvr"
	"go.uber.org/zap"
	"strings"
)

func main() {
	config, err := resolvr.LoadConfig()
	if err != nil {
		zap.S().Panic("error loading config", "error", err.Error())
	}

	var logger *zap.Logger
	if strings.EqualFold(config.Env, "prod") {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}

	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	resolvr.ServeDns(config)
}
