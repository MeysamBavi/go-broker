package config

import "time"

type Scheduler struct {
	StartRPS        float64
	TargetRPS       float64
	RiseDuration    time.Duration
	PlateauDuration time.Duration
	Subjects        int
}
