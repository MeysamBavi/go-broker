package metrics

type Handler interface {
	IncPublishCallCount(success bool)
	IncSubscribeCallCount(success bool)
	IncFetchCallCount(success bool)
	ReportPublishLatency(value float64)
	ReportSubscribeLatency(value float64)
	ReportFetchLatency(value float64)
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

func (n noImpl) ReportPublishLatency(_ float64) {}

func (n noImpl) ReportSubscribeLatency(_ float64) {}

func (n noImpl) ReportFetchLatency(_ float64) {}

func (n noImpl) IncActiveSubscribers() {}

func (n noImpl) DecActiveSubscribers() {}
