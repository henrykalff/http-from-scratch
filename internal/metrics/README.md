# Metrics
This repository enables the measuring of metrics with VictoriaMetrics.

## Quickstart
Pull the VictoriaMetrics Docker image:
```
docker pull victoriametrics/victoria-metrics:v1.155.0
```

Start a local container and expose at Port 8248:
```
docker run -it --rm -v `pwd`/victoria-metrics-data:/victoria-metrics-data -p 8428:8428 victoriametrics/victoria-metrics:v1.155.0 --selfScrapeInterval=5s -storageDataPath=victoria-metrics-data
```