package cm600exporter

import (
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = &Exporter{}

// Exporter exports prometheus metrics for a CM600 modem
type Exporter struct {
	mu     sync.Mutex
	client *client

	m *ModemCollector
}

// New creates a new CM600 Exporter Instance
func New(httpClient *http.Client, url, username, password string) (*Exporter, error) {
	cl := &client{client: httpClient, url: url, username: username, password: password}
	modemCollector := NewModemCollector(cl)
	return &Exporter{client: cl, m: modemCollector}, nil
}

// Collect gathers the statistics from the specified modem
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	if err := e.m.collect(); err != nil {
		log.Println("failed to collect modem metrics:", err)
	}

	for _, metric := range e.m.collectorList() {
		metric.Collect(ch)
	}
}

// Describe describes the metrics available for this modem
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.m.describe(ch)
}
