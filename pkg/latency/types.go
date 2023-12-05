package latency

import "fmt"

type Destination struct {
	Ip          string
	CountryCode string
}

type Latency struct {
	LatencyMeanMs   float64
	LatencyMedianMs float64
	LatencyP90Ms    float64
	PacketLoss      float64
}

type Result struct {
	Ip            string
	CountryCode   string
	AlternativeIP string
	ErrorMessage  string
	Successful    bool
	Latency       Latency
}

func (r *Result) GetResult() string {
	return fmt.Sprintf("Successful: %t \nDestination: %s \nCountry code: %s \nAlternate destination: %s \nMedian latency: %.1fms", r.Successful, r.Ip, r.CountryCode, r.AlternativeIP, r.Latency.LatencyMedianMs)
}
