package main

import (
	"github.com/RileyMichael/resolvr/internal/resolvr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
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

	// todo: extract
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(config.MetricsAddress, nil)
	}()

	resolvr.ServeDns(config)
}
