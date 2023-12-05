package latency

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/max-rocket-internet/country-latency-metronomy/pkg/ipcountry"
	"github.com/montanaflynn/stats"
	probing "github.com/prometheus-community/pro-bing"
	log "github.com/sirupsen/logrus"
)

func GetLatency(destination Destination) (result Result, err error) {
	result.Ip = destination.Ip
	reachable, err := canPing(destination.Ip)

	if err != nil {
		return result, errors.New(fmt.Sprintf("Ping error to '%s': %s \n", destination.Ip, err.Error()))
	}

	if destination.CountryCode == "" {
		destination.CountryCode, err = ipcountry.GetIpCountry(destination.Ip)
	}

	result.CountryCode = strings.ToLower(destination.CountryCode)

	if err != nil && !reachable {
		return result, errors.New(fmt.Sprintf("destination is not reachable and whois lookup failed for '%s': %s", destination.Ip, err.Error()))
	}

	if reachable {
		latency, err := measureLatency(destination.Ip)
		if err != nil {
			return result, err
		}

		result.Successful = true
		result.Latency = latency

		return result, nil
	}

	latency, alternateIp, err := getAlternativeLatency(destination)
	if err != nil {
		return result, err
	}

	result.Latency = latency
	result.AlternativeIP = alternateIp
	result.Successful = true

	return result, nil
}

func getAlternativeLatency(d Destination) (result Latency, alternateIp string, err error) {
	tracerouteHops, err := tracerouteIp(d.Ip)
	if err != nil {
		return result, "", errors.New(fmt.Sprintf("traceroute error for '%s': %s", d.Ip, err.Error()))
	}

	alternateIp, err = filterTraceRouteHops(d.CountryCode, tracerouteHops)
	if err != nil {
		return result, "", errors.New(fmt.Sprintf("traceroute filtering error for '%s': %s", d.Ip, err.Error()))
	}

	latency, err := measureLatency(alternateIp)
	if err != nil {
		return result, "", err
	}

	return latency, alternateIp, nil
}

func canPing(host string) (result bool, err error) {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Could not create Pinger to '%s'", host))
	}

	pinger.Count = 1
	pinger.Timeout = time.Duration(time.Second * 2)

	err = pinger.Run()
	if err != nil {
		return false, errors.New(fmt.Sprintf("Could not run Pinger to '%s' \n", host))
	}

	pingStats := pinger.Statistics()
	if pingStats.PacketsRecv == 1 {
		log.Debugf("Ping test successful to '%s': %d avg\n", host, pingStats.AvgRtt.Milliseconds())
		return true, nil
	}

	log.Debugf("Ping failed to '%s', no reply \n", host)

	return false, nil
}

func measureLatency(ip string) (result Latency, err error) {
	log.Debugf("Starting latancy analysis of '%s' \n", ip)

	pinger, err := probing.NewPinger(ip)
	if err != nil {
		return result, errors.New(fmt.Sprintf("error creating pinger: %s", err.Error()))
	}

	pinger.Count = 10
	pinger.Timeout = time.Duration(time.Second * 10)

	err = pinger.Run()
	if err != nil {
		return result, errors.New(fmt.Sprintf("error running pinger: %s", err.Error()))
	}

	pingStats := pinger.Statistics()

	if pingStats.PacketLoss > 50 {
		return result, errors.New(fmt.Sprintf("packet loss too high for '%s': %.1f%%", ip, pingStats.PacketLoss))
	}

	if len(pingStats.Rtts) < 4 {
		return result, errors.New(fmt.Sprintf("too few Rtts for '%s': %d", ip, len(pingStats.Rtts)))
	}

	pingMsRtts := []float64{}
	for _, i := range pingStats.Rtts {
		pingMsRtts = append(pingMsRtts, float64(i.Milliseconds()))
	}

	mean, err := stats.Mean(pingMsRtts)
	if err != nil {
		return result, errors.New(fmt.Sprintf("error calculating mean: %s", err.Error()))
	}

	median, err := stats.Median(pingMsRtts)
	if err != nil {
		return result, errors.New(fmt.Sprintf("error calculating median: %s", err.Error()))
	}

	p90, err := stats.Percentile(pingMsRtts, 90)
	if err != nil {
		return result, errors.New(fmt.Sprintf("error calculating percentile: %s", err.Error()))
	}

	result.LatencyMedianMs = median
	result.LatencyMeanMs = mean
	result.LatencyP90Ms = p90
	result.PacketLoss = pingStats.PacketLoss

	return result, nil
}
