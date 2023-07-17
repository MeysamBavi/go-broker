package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
	"time"
)

const (
	publish      = "publish"
	subscribe    = "subscribe"
	fetch        = "fetch"
	successLabel = "success"
	methodLabel  = "method"
)

type prometheusImpl struct {
	methodCount       *prometheus.CounterVec
	methodDuration    *prometheus.SummaryVec
	activeSubscribers *prometheus.GaugeVec
}

func NewPrometheusHandler() Handler {
	return &prometheusImpl{
		methodCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "method_count",
			Help: "number of failed/successful calls for each rpc endpoint",
		}, []string{successLabel, methodLabel}),
		methodDuration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: "method_duration",
			Help: "the method latency for each rpc endpoint in milliseconds",
			Objectives: map[float64]float64{
				.99: .01,
				.95: .01,
				.50: .01,
			},
		}, []string{methodLabel}),
		activeSubscribers: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "active_subscribers",
			Help: "number of active subscribers",
		}, []string{}),
	}
}

func (p *prometheusImpl) incMethodCount(method string, success bool) {
	p.methodCount.
		With(prometheus.Labels{methodLabel: method, successLabel: strconv.FormatBool(success)}).
		Inc()
}

func (p *prometheusImpl) reportMethodLatency(method string, latency time.Duration) {
	p.methodDuration.
		With(prometheus.Labels{methodLabel: method}).
		Observe(float64(latency.Milliseconds()))
}

func (p *prometheusImpl) IncPublishCallCount(success bool) {
	p.incMethodCount(publish, success)
}

func (p *prometheusImpl) IncSubscribeCallCount(success bool) {
	p.incMethodCount(subscribe, success)
}

func (p *prometheusImpl) IncFetchCallCount(success bool) {
	p.incMethodCount(fetch, success)
}

func (p *prometheusImpl) ReportPublishLatency(value time.Duration) {
	p.reportMethodLatency(publish, value)
}

func (p *prometheusImpl) ReportFetchLatency(value time.Duration) {
	p.reportMethodLatency(fetch, value)
}

func (p *prometheusImpl) IncActiveSubscribers() {
	p.activeSubscribers.
		With(prometheus.Labels{}).Inc()
}

func (p *prometheusImpl) DecActiveSubscribers() {
	p.activeSubscribers.
		With(prometheus.Labels{}).Dec()
}
