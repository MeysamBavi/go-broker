package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/MeysamBavi/go-broker/api/client/collector"
	"github.com/MeysamBavi/go-broker/api/client/config"
	"github.com/MeysamBavi/go-broker/api/client/scheduler"
	"github.com/MeysamBavi/go-broker/api/client/sender"
	"log"
)

var (
	host     = flag.String("host", "localhost:50043", "the host to connect to")
	verbose  = flag.Bool("verbose", false, "log more info")
	subjects = flag.Int("subjects", 0, "limits the number of different subjects used")
)

func main() {
	flag.Parse()

	cfg := config.Load()
	cfgJson, err := json.MarshalIndent(cfg, "", "	")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", cfgJson)

	s := sender.Sender{
		Host: *host,
		PublishStream: scheduler.SchedulePublish(config.Scheduler{
			StartRPS:        cfg.StartRPS,
			TargetRPS:       cfg.TargetRPS,
			RiseDuration:    cfg.RiseDuration,
			PlateauDuration: cfg.PlateauDuration,
			Subjects:        *subjects,
		}),
		Connections: cfg.Connections,
		Verbose:     *verbose,
	}

	for summary := range collector.Collect(s.Start()) {
		fmt.Printf("perceived throughput: %8.2f\terror rate: %.2f\n", summary.Throughput, summary.ErrorRate)
		if summary.ErrorRate > cfg.ErrorRateThreshold {
			log.Fatal("error rate passed threshold")
		}
	}
}
