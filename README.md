##### Rounds
Simple consensus algorithms playground

Generate certs
```
./scripts/makecerts.sh certs *your_email*
```

Run nodes & storage
```
make build
make run
```

#### Telemetry
[OpenCensus](https://opencensus.io/introduction/)

[Metrics](http://localhost:9090/)
```
./scripts/metrics.sh
```

[Tracing](http://localhost:16686)
```
./scripts/tracing.sh
```
