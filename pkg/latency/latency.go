package latency

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/max-rocket-internet/country-latency-metronomy/pkg/ipcountry"
	"github.com/montanaflynn/stats"
	probing "github.com/prometheus-community/pro-bing"
	log "github.com/sirupsen/logrus"
)

func GetLatency(destination Destination) (result Result, err error) {
	err = validateDestination(destination)
	if err != nil {
		return result, err
	}

	result.Ip = destination.Ip

	reachable, err := canPing(destination.Ip)
	if err != nil {
		return result, fmt.Errorf("Ping error to '%s': %s \n", destination.Ip, err.Error())
	}

	if destination.CountryCode == "" {
		destination.CountryCode, err = ipcountry.GetIpCountry(destination.Ip)
	} else {
		destination.CountryCode = strings.ToLower(destination.CountryCode)
	}

	if err != nil && !reachable {
		return result, fmt.Errorf("destination is not reachable and whois lookup failed for '%s': %s", destination.Ip, err.Error())
	}

	result.CountryCode = destination.CountryCode

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
		return result, "", fmt.Errorf("traceroute error for '%s': %s", d.Ip, err.Error())
	}

	alternateIp, err = filterTraceRouteHops(d.CountryCode, tracerouteHops)
	if err != nil {
		return result, "", fmt.Errorf("traceroute filtering error for '%s': %s", d.Ip, err.Error())
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
		return false, fmt.Errorf("Could not create Pinger to '%s'", host)
	}

	pinger.Count = 1
	pinger.Timeout = time.Duration(time.Second * 2)

	err = pinger.Run()
	if err != nil {
		return false, fmt.Errorf("Could not run Pinger to '%s' \n", host)
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
		return result, fmt.Errorf("error creating pinger: %s", err.Error())
	}

	pinger.Count = 10
	pinger.Timeout = time.Duration(time.Second * 10)

	err = pinger.Run()
	if err != nil {
		return result, fmt.Errorf("error running pinger: %s", err.Error())
	}

	pingStats := pinger.Statistics()

	if pingStats.PacketLoss > 50 {
		return result, fmt.Errorf("packet loss too high for '%s': %.1f%%", ip, pingStats.PacketLoss)
	}

	if len(pingStats.Rtts) < 4 {
		return result, fmt.Errorf("too few Rtts for '%s': %d", ip, len(pingStats.Rtts))
	}

	pingMsRtts := []float64{}
	for _, i := range pingStats.Rtts {
		pingMsRtts = append(pingMsRtts, float64(i.Milliseconds()))
	}

	mean, err := stats.Mean(pingMsRtts)
	if err != nil {
		return result, fmt.Errorf("error calculating mean: %s", err.Error())
	}

	median, err := stats.Median(pingMsRtts)
	if err != nil {
		return result, fmt.Errorf("error calculating median: %s", err.Error())
	}

	p90, err := stats.Percentile(pingMsRtts, 90)
	if err != nil {
		return result, fmt.Errorf("error calculating percentile: %s", err.Error())
	}

	result.MedianMs = median
	result.MeanMs = mean
	result.P90Ms = p90
	result.PacketLoss = pingStats.PacketLoss

	return result, nil
}

func validateDestination(destination Destination) error {
	if destination.Ip == "" {
		return errors.New("Destination IP is blank")
	}

	if destination.CountryCode != "" && len(destination.CountryCode) != 2 {
		return errors.New("Country code should be 2 characters (ISO 3166-1 alpha-2)")
	}

	destinationIp, err := netip.ParseAddr(destination.Ip)
	if err != nil {
		return err
	}
	if destinationIp.IsPrivate() {
		return fmt.Errorf("Private IP address (RFC 1918) given '%s'", destination.Ip)
	}
	if destinationIp.IsLoopback() {
		return fmt.Errorf("Loopback address given '%s'", destination.Ip)
	}
	if destinationIp.IsMulticast() {
		return fmt.Errorf("Multicast address given '%s'", destination.Ip)
	}

	return nil
}
