package main

import (
	"github.com/RileyMichael/resolvr/internal/resolvr"
	"go.uber.org/zap"
)

func main() {
	// init logger.. for now, just default to dev mode
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	config, err := resolvr.LoadConfig()

	if err != nil {
		zap.S().Panic("error loading config", "error", err.Error())
	}

	resolvr.ServeDns(config)
}
