package latency

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"sync"

	"github.com/aeden/traceroute"
	"github.com/max-rocket-internet/country-latency-metronomy/pkg/ipcountry"
	log "github.com/sirupsen/logrus"
)

type traceRouteHops struct {
	hops []string
	sync.RWMutex
}

func (i *traceRouteHops) Read() []string {
	i.RLock()
	defer i.RUnlock()
	return i.hops
}

func (i *traceRouteHops) Length() int {
	i.RLock()
	defer i.RUnlock()
	return len(i.hops)
}

func (i *traceRouteHops) Append(hop string) {
	i.Lock()
	defer i.Unlock()
	i.hops = append(i.hops, hop)
}

func tracerouteIp(ip string) ([]string, error) {
	log.Debugf("Starting traceroute to '%s' \n", ip)
	options := traceroute.TracerouteOptions{}
	options.SetMaxHops(24)
	options.SetTimeoutMs(500)

	c := make(chan traceroute.TracerouteHop, 0)
	hops := traceRouteHops{}

	go func() {
		for {
			hop, ok := <-c
			if !ok {
				return
			}
			if hop.Success {
				addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
				hops.Append(addr)
			}
		}
	}()

	_, err := traceroute.Traceroute(ip, &options, c)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error doing trceroute: %s", err.Error()))
	}

	log.Debugf("Traceroute to '%s' finished, %d hops \n", ip, hops.Length())

	return hops.Read(), err
}

func filterTraceRouteHops(country string, traceRouteHops []string) (result string, err error) {
	log.Debugf("Filtering traceroute results for country code '%s' \n", country)

	slices.Reverse(traceRouteHops)

	for _, ip := range traceRouteHops {
		netIp := net.ParseIP(ip)

		if netIp.IsPrivate() {
			log.Debugf("Skipping '%s' as it's a private address \n", ip)
			continue
		}

		hopCountry, err := ipcountry.GetIpCountry(ip)
		if err != nil {
			log.Debugf("Skipping '%s' as whois failed: %s \n", ip, err.Error())
			continue
		}

		if hopCountry != country {
			log.Debugf("Skipping '%s' as country '%s' does not match '%s' \n", ip, hopCountry, country)
			continue
		}

		log.Debugf("Found latest good hop '%s' \n", ip)

		return ip, nil
	}

	return result, errors.New(fmt.Sprintf("No suitable hops found for country '%s'", country))
}
