package main

import (
	"github.com/spf13/viper"
	"rounds/ledger"
)

func main() {
	ledger.LoadConfig()
	cfg := &ledger.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		panic(err)
	}

	go ledger.Serve(cfg)
	ledger.PromExporter(cfg)
	ledger.Tracing(cfg)

	select {}
}
