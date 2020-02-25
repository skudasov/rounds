package main

import (
	"go.opencensus.io/stats/view"
	"rounds/node"
	"rounds/telemetry"
)

func main() {
	cfg := node.MakeConfig()
	node.ValidateConfig(cfg)

	node.WriteKeyPairIfNotExists(cfg)
	priv, pub, pubPem := node.LoadKeyPair(cfg)
	n := node.NewNode(cfg, priv, pub, pubPem)

	// Receive all messages, switch by type
	go n.StartTransport()
	// Sync rounds between nodes and schedule consensus start
	go n.Schedule(cfg)
	// Loop consensus rounds
	go n.Processing()

	telemetry.PromExporter(cfg.Opencensus)
	telemetry.Tracing(cfg.Opencensus)
	if err := view.Register(node.LatencyView); err != nil {
		panic(err)
	}
	telemetry.ServeZPages(cfg.Opencensus)

	select {}
}
