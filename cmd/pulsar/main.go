package main

import (
	"go.opencensus.io/stats/view"
	"rounds/node"
)

func main() {
	cfg := node.MakeConfig()
	node.ValidateConfig(cfg)

	node.WriteKeyPairIfNotExists(cfg)
	priv, pub, pubPem := node.LoadKeyPair(cfg)
	n := node.NewNode(cfg, priv, pub, pubPem)

	// Connect to peers, reconnect if conn is nil
	go n.ConnectPeers()
	// Receive all messages, switch by type
	go n.ReceiveLoop()
	// Sync rounds between nodes and schedule consensus start
	go n.Schedule(cfg)
	// Loop consensus rounds
	go n.Processing()

	node.PromExporter(cfg)
	node.Tracing(cfg)
	if err := view.Register(node.LatencyView); err != nil {
		panic(err)
	}

	select {}
}
