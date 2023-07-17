package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
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

func (p *prometheusImpl) reportMethodLatency(method string, latency float64) {
	p.methodDuration.
		With(prometheus.Labels{methodLabel: method}).
		Observe(latency)
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

func (p *prometheusImpl) ReportPublishLatency(value float64) {
	p.reportMethodLatency(publish, value)
}

func (p *prometheusImpl) ReportSubscribeLatency(value float64) {
	p.reportMethodLatency(subscribe, value)
}

func (p *prometheusImpl) ReportFetchLatency(value float64) {
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
