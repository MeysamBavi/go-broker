package metrics

import "time"

type Handler interface {
	IncPublishCallCount(success bool)
	IncSubscribeCallCount(success bool)
	IncFetchCallCount(success bool)
	ReportPublishLatency(value time.Duration)
	ReportSubscribeLatency(value time.Duration)
	ReportFetchLatency(value time.Duration)
	IncActiveSubscribers()
	DecActiveSubscribers()
}

type noImpl struct{}

func NewEmptyHandler() Handler {
	return noImpl{}
}

func (n noImpl) IncPublishCallCount(_ bool) {}

func (n noImpl) IncSubscribeCallCount(_ bool) {}

func (n noImpl) IncFetchCallCount(_ bool) {}

func (n noImpl) ReportPublishLatency(_ time.Duration) {}

func (n noImpl) ReportSubscribeLatency(_ time.Duration) {}

func (n noImpl) ReportFetchLatency(_ time.Duration) {}

func (n noImpl) IncActiveSubscribers() {}

func (n noImpl) DecActiveSubscribers() {}
