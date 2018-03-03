package cm600exporter

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "cm600"

var (
	labelsDownstream = []string{"downstream", "dcid"}
	labelsUpstream   = []string{"upstream", "ucid"}
)

// ModemCollector represents a single group of stats
type ModemCollector struct {
	c *client

	// Downstream metrics
	DownstreamFreq *prometheus.GaugeVec

	DownstreamPower *prometheus.GaugeVec

	DownstreamSNR *prometheus.GaugeVec

	DownstreamModulation *prometheus.GaugeVec

	DownstreamCorrecteds *prometheus.GaugeVec

	DownstreamUncorrectables *prometheus.GaugeVec

	// Upstrem Metrics

	UpstreamFreq *prometheus.GaugeVec

	UpstreamPower *prometheus.GaugeVec

	UpstreamSymbolRate *prometheus.GaugeVec
}

// NewModemCollector creates a new statistics collector
func NewModemCollector(c *client) *ModemCollector {
	return &ModemCollector{
		c: c,

		DownstreamFreq: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "downstrem_freq_hertz",
				Help:      "Modem Downstream Frequency (Hz)",
			},
			labelsDownstream,
		),

		DownstreamPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "downstream_power_dbmv",
				Help:      "Modem Downstream Power (dBmV)",
			},
			labelsDownstream,
		),

		DownstreamSNR: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "downstream_snr_db",
				Help:      "Modem Downstream SNR (dB)",
			},
			labelsDownstream,
		),

		DownstreamModulation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "downstream_modulation_qam",
				Help:      "Modem Downstream Modulation (QAM)",
			},
			labelsDownstream,
		),

		DownstreamCorrecteds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "dowmstream_correcteds_total",
				Help:      "Modem Downstream Correcteds",
			},
			labelsDownstream,
		),

		DownstreamUncorrectables: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "downstream_uncorrectables_total",
				Help:      "Modem Downstream Uncorrectables",
			},
			labelsDownstream,
		),

		UpstreamFreq: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "upstream_freq_hertz",
				Help:      "Modem Upstream Frequency (MHz)",
			},
			labelsUpstream,
		),

		UpstreamPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "upstream_power_dbmv",
				Help:      "Modem Upstream Power (dBmV)",
			},
			labelsUpstream,
		),

		UpstreamSymbolRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "upstream_symbol_rate",
				Help:      "Modem Upstream Symbol Rate (kSym/s)",
			},
			labelsUpstream,
		),
	}
}

func (m *ModemCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{
		m.DownstreamCorrecteds,
		m.DownstreamSNR,
		m.DownstreamModulation,
		m.DownstreamUncorrectables,
		m.DownstreamFreq,
		m.DownstreamPower,
		m.UpstreamFreq,
		m.UpstreamSymbolRate,
		m.UpstreamPower,
	}
}

func (m *ModemCollector) collect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := m.c.fetch(ctx)
	if err != nil {
		return err
	}
	for id, node := range data.ds {
		downstreamID := strconv.Itoa(id + 1)
		downstreamDCID := strconv.Itoa(node.DCID)
		m.DownstreamSNR.WithLabelValues(downstreamID, downstreamDCID).Set(node.SNR)
		m.DownstreamFreq.WithLabelValues(downstreamID, downstreamDCID).Set(node.Freq)
		m.DownstreamPower.WithLabelValues(downstreamID, downstreamDCID).Set(node.Power)
		m.DownstreamModulation.WithLabelValues(downstreamID, downstreamDCID).Set(float64(node.Modulation))
		m.DownstreamCorrecteds.WithLabelValues(downstreamID, downstreamDCID).Set(float64(node.Correcteds))
		m.DownstreamUncorrectables.WithLabelValues(downstreamID, downstreamDCID).Set(float64(node.Uncorrectables))
	}

	for id, node := range data.us {
		upstreamID := strconv.Itoa(id + 1)
		upstreamUCID := strconv.Itoa(node.UCID)

		m.UpstreamFreq.WithLabelValues(upstreamID, upstreamUCID).Set(node.Freq)
		m.UpstreamPower.WithLabelValues(upstreamID, upstreamUCID).Set(node.Power)
		m.UpstreamSymbolRate.WithLabelValues(upstreamID, upstreamUCID).Set(float64(node.SymbolRate))
	}

	return nil
}

func (m *ModemCollector) describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.collectorList() {
		metric.Describe(ch)
	}
}
