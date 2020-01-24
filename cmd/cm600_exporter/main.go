package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	cm600exporter "github.com/nickvanw/cm600_exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsPath = flag.String("metrics.path", "/metrics", "path to fetch metrics")
	metricsAddr = flag.String("metrics.addr", ":9191", "address to listen")

	modemUsername = flag.String("modem.username", "admin", "username to get stats")
	modemPassword = flag.String("modem.password", "password", "password for modem")
	modemAddr     = flag.String("modem.host", "http://192.168.100.1/DocsisStatus.asp", "url to fetch modem")
)

func main() {
	flag.Parse()

	c, err := createExporter(*modemAddr, *modemUsername, *modemPassword)
	if err != nil {
		log.Fatalf("unable to create client: %s", err)
	}

	prometheus.MustRegister(c)

	mux := http.NewServeMux()
	mux.Handle(*metricsPath, promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Netgear CM600 Exporter</title></head>
			<body>
			<h1>Netgear CM600 Exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})
	loggedMux := handlers.LoggingHandler(os.Stdout, mux)
	if err := http.ListenAndServe(*metricsAddr, loggedMux); err != nil {
		log.Fatalf("unable to start metrics server: %s", err)
	}
}

func createExporter(modem, login, password string) (*cm600exporter.Exporter, error) {
	client := http.Client{}
	return cm600exporter.New(&client, modem, login, password)
}
