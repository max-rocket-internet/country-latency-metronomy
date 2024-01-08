package latency

import "fmt"

type Destination struct {
	Ip          string
	CountryCode string
}

type Latency struct {
	MeanMs     float64
	MedianMs   float64
	P90Ms      float64
	PacketLoss float64
}

type Result struct {
	Ip            string
	CountryCode   string
	AlternativeIP string
	ErrorMessage  string
	Successful    bool
	Latency       Latency
}

func (r *Result) GenerateText() string {
	return fmt.Sprintf("Successful: %t \nDestination: %s \nCountry code: %s \nAlternate destination: %s \nMedian latency: %.1fms", r.Successful, r.Ip, r.CountryCode, r.AlternativeIP, r.Latency.MedianMs)
}

func (r *Result) GenerateCsv(includeHeader bool) (result string) {
	result = fmt.Sprintf("%s,%.1f,%s,%s,%s", r.Ip, r.Latency.MedianMs, r.AlternativeIP, r.CountryCode, r.ErrorMessage)

	if includeHeader {
		result = "Ip,LatencyMedianMs,AlternativeIP,CountryCode,ErrorMessage" + "\n" + result
	}

	return result
}
