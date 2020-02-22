package node

import (
	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
	"log"
	"net/http"
	"time"
)

var (
	BroadcastMs = stats.Float64("broadcast/latency", "The latency in milliseconds per broadcast", "ms")

	LatencyView = &view.View{
		Name:        "broadcast/latency",
		Measure:     BroadcastMs,
		Description: "broadcast latencies distribution",

		// Latency in buckets:
		// [>=0ms, >=25ms, >=50ms, >=75ms, >=100ms, >=200ms, >=400ms, >=600ms, >=800ms, >=1s, >=2s, >=4s, >=6s]
		Aggregation: view.Distribution(0, 25, 50, 75, 100, 200, 400, 600, 800, 1000, 2000, 4000, 6000),
		TagKeys:     []tag.Key{KeyMethod, KeyLabel}}
)

var (
	KeyLabel, _  = tag.NewKey("node")
	KeyMethod, _ = tag.NewKey("method")
	KeyStatus, _ = tag.NewKey("status")
	KeyError, _  = tag.NewKey("error")
)

func SinceInMilliseconds(startTime time.Time) float64 {
	return float64(time.Since(startTime).Nanoseconds()) / 1e6
}

func ServeZPages(c *Config) {
	mux := http.NewServeMux()
	zpages.Handle(mux, "/debug")

	port := ":" + c.Opencensus.ZPages.Port
	log.Printf("serving zPages on port: %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Failed to serve zPages")
	}
}

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
