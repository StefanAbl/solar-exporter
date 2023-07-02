package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"solar-exporter/collector"
)

var (
	addr = flag.String("listen-address", ":9090", "The address to listen on for HTTP requests.")

	url  = os.Getenv("URL")
	user = os.Getenv("USER")
	pw   = os.Getenv("PW")
)

func main() {

	var r = prometheus.NewRegistry()
	var collector = collector.NewCollector(url, user, pw)
	r.MustRegister(collector)
	flag.Parse()
	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
