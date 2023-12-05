package main

import (
	"flag"
	"fmt"

	"github.com/max-rocket-internet/country-latency-metronomy/pkg/latency"
	log "github.com/sirupsen/logrus"
)

func main() {
	debugLogging := flag.Bool("debug", false, "Enable debug logging")
	ip := flag.String("ip-address", "", "IP address to measure latency to")

	flag.Parse()

	if *ip == "" {
		log.Fatalln("Must set -ip-address flag")
	}

	if *debugLogging {
		log.SetLevel(log.DebugLevel)
	}

	destination := latency.Destination{Ip: *ip, CountryCode: ""}

	result, err := latency.GetLatency(destination)
	if err != nil {
		log.Error(err)
	}

	fmt.Println(result.GetResult())
}
