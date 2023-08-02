package collector

import (
	"github.com/MeysamBavi/go-broker/api/client/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type ResponseLog struct {
	At    time.Time
	Error error
}

type Summary struct {
	Throughput float64
	ErrorRate  float64
}

const (
	reportPeriod = time.Second
	samples      = 1000
)

func Collect(cfg config.Collector, logStream <-chan ResponseLog) <-chan Summary {
	index := 0
	buffer := make([]ResponseLog, samples)

	failed := 0
	total := 0
	summaryStream := make(chan Summary, 5)
	lastReport := time.Now()
	report := func() {
		lastReport = time.Now()
		minAt := lastReport
		for i := 0; i < index; i++ {
			log := buffer[i]
			if log.At.Before(minAt) {
				minAt = log.At
			}
			if log.Error != nil && (!cfg.IgnoreUnavailableError || status.Code(log.Error) != codes.Unavailable) {
				failed++
			}
			total++
		}
		summaryStream <- Summary{
			Throughput: float64(index) / time.Since(minAt).Seconds(),
			ErrorRate:  float64(failed) / float64(total),
		}
		index = 0
	}

	go func() {
		ticker := time.NewTicker(reportPeriod)
		defer ticker.Stop()
		defer close(summaryStream)
		for {
			select {
			case currentTime := <-ticker.C:
				if currentTime.Sub(lastReport) >= reportPeriod {
					report()
				}
			case log, ok := <-logStream:
				if !ok {
					report()
					return
				}
				buffer[index] = log
				index++
				if index >= len(buffer) {
					report()
				}
			}
		}
	}()

	return summaryStream
}
