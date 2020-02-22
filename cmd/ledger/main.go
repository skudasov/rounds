package main

import (
	"github.com/spf13/viper"
	"rounds/ledger"
	"rounds/telemetry"
)

func main() {
	ledger.LoadConfig()
	cfg := &ledger.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		panic(err)
	}

	go ledger.Serve(cfg)
	telemetry.PromExporter(cfg.Opencensus)
	telemetry.Tracing(cfg.Opencensus)
	telemetry.ServeZPages(cfg.Opencensus)

	select {}
}
