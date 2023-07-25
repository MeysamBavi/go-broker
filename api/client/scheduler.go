package main

import (
	"context"
	"github.com/MeysamBavi/go-broker/api/client/config"
	pb "github.com/MeysamBavi/go-broker/api/proto"
	"math"
	"sync"
	"time"
)

const (
	channelBuffer = 100
	timeStep      = time.Microsecond
)

var subjects = []string{
	"alpha",
	"beta",
	"gamma",
	"delta",
	"epsilon",
	"zeta",
	"eta",
	"theta",
	"iota",
	"kappa",
}

func SchedulePublish(cfg config.Scheduler) <-chan *pb.PublishRequest {
	supplier := func() *pb.PublishRequest {
		return &pb.PublishRequest{
			Subject:           RandomItem(subjects) + RandomStringWithCharset(4, "01"),
			Body:              []byte(RandomString(RandomInt(10, 30))),
			ExpirationSeconds: int32(RandomInt(10, 20)),
		}
	}
	return schedule(cfg, supplier)
}

func schedule[T any](cfg config.Scheduler, supplier func() T) <-chan T {
	ch := make(chan T, channelBuffer)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		riseGoalFunc := getSentRequestsCountFunction(cfg.StartRPS, cfg.TargetRPS, cfg.RiseDuration)
		riseCtx, riseCancel := context.WithTimeout(context.Background(), cfg.RiseDuration)
		defer riseCancel()
		wg.Done()
		send(riseCtx, ch, riseGoalFunc, supplier)

		plateauGoalFunc := getSentRequestsCountFunction(cfg.TargetRPS, cfg.TargetRPS, cfg.PlateauDuration)
		plateauCtx, plateauCancel := context.WithTimeout(context.Background(), cfg.PlateauDuration)
		defer plateauCancel()
		send(plateauCtx, ch, plateauGoalFunc, supplier)

		close(ch)
	}()

	wg.Wait()
	return ch
}

func send[T any](ctx context.Context, ch chan<- T, goalFunc func(t float64) float64, supplier func() T) float64 {
	ticker := time.NewTicker(timeStep)

	achievedGoal := 0.0
	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return achievedGoal
		case currentTime := <-ticker.C:
			t := currentTime.Sub(startTime).Seconds()
			goal := goalFunc(t)
			toBeAdded := math.Max(0, math.Round(goal-achievedGoal))
			for i := 0; i < int(toBeAdded); i++ {
				ch <- supplier()
			}
			achievedGoal += toBeAdded
		}
	}
}

func getSentRequestsCountFunction(start, target float64, d time.Duration) func(t float64) float64 {
	// rate:
	// r(t) = start + ((target - start) / d) * t
	// anti-derivative of rate:
	// c(t) = start * t + 0.5 * ((target - start) / d) * t * t

	return func(t float64) float64 {
		return start*t + 0.5*((target-start)/d.Seconds())*t*t
	}
}
