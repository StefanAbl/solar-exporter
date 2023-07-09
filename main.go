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

func main() {
	var (
		addr = flag.String("listen-address", ":9090", "The address to listen on for HTTP requests.")
		url  = os.Getenv("URL")
		user = os.Getenv("USER")
		pw   = os.Getenv("PASS")
	)
	flag.Parse()
	server := createServer(addr, url, user, pw)
	log.Fatal(server.ListenAndServe())
}

func createServer(addr *string, url string, user string, pw string) http.Server {
	r := prometheus.NewRegistry()
	var collector = collector.NewCollector(url, user, pw)
	r.MustRegister(collector)

	var mux = http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

	return http.Server{
		Handler: mux,
		Addr:    *addr,
	}

}
