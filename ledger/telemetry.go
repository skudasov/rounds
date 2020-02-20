package ledger

import (
	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/trace"
	"log"
	"net/http"
)

func Tracing(c *Config) {
	exporter, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: ":" + c.Opencensus.Jaeger.Port,
		Process: jaeger.Process{
			ServiceName: c.Opencensus.Jaeger.Nodelabel,
			Tags: []jaeger.Tag{
				jaeger.StringTag("hostname", "localhost"),
			},
		},
	})
	if err != nil {
		return
	}
	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})
	defer exporter.Flush()
}

func PromExporter(cfg *Config) {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "rounds",
	})
	if err != nil {
		log.Fatalf("Failed to create the Prometheus stats exporter: %v", err)
	}
	port := cfg.Opencensus.Prometheus.Port
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		log.Printf("prometheus metrics on %s", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Fatalf("Failed to run Prometheus scrape endpoint: %v", err)
		}
	}()
}
