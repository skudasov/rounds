package telemetry

import (
	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
	"log"
	"net/http"
)

type OpencensusConfig struct {
	Prometheus struct {
		Nodelabel string `validate:"required"`
		Port      string `validate:"required"`
	} `validate:"required"`
	Jaeger struct {
		Nodelabel string `validate:"required"`
		Port      string `validate:"required"`
	} `validate:"required"`
	ZPages struct {
		Port string `json:"port",validate:"required"`
	} `json:"zpages",validate:"required"`
}

func ServeZPages(c OpencensusConfig) {
	mux := http.NewServeMux()
	zpages.Handle(mux, "/debug")

	port := ":" + c.ZPages.Port
	log.Printf("serving zPages on port: %s", port)
	go func() {
		if err := http.ListenAndServe(port, mux); err != nil {
			log.Fatalf("Failed to serve zPages")
		}
	}()
}

func Tracing(c OpencensusConfig) {
	exporter, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: ":" + c.Jaeger.Port,
		Process: jaeger.Process{
			ServiceName: c.Jaeger.Nodelabel,
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

func PromExporter(cfg OpencensusConfig) {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "rounds",
	})
	if err != nil {
		log.Fatalf("Failed to create the Prometheus stats exporter: %v", err)
	}
	port := cfg.Prometheus.Port
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		log.Printf("prometheus metrics on %s", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Fatalf("Failed to run Prometheus scrape endpoint: %v", err)
		}
	}()
}
