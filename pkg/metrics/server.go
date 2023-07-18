package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func RunServer(config Config) {
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("metrics http server listening on port %s\n", config.HttpPort)
	if err := http.ListenAndServe(":"+config.HttpPort, nil); err != nil {
		log.Fatal("could not server metrics: ", err)
	}
}
