package ipcountry

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/digineo/ripego"
)

var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	cache                = whoisIpCountryCache{}
)

type whoisIpCountryCache struct {
	countries map[string][]string
	sync.RWMutex
}

func (i *whoisIpCountryCache) Get(ipAddress string) string {
	i.RLock()
	defer i.RUnlock()

	for c, ips := range i.countries {
		for _, ip := range ips {
			if ip == ipAddress {
				return c
			}
		}
	}

	return ""
}

func (i *whoisIpCountryCache) Save(countryCode string, ipAddress string) {
	i.Lock()
	defer i.Unlock()

	if !slices.Contains(i.countries[countryCode], ipAddress) {
		i.countries[countryCode] = append(i.countries[countryCode], ipAddress)
	}
}

func init() {
	cache.countries = make(map[string][]string)
}

func GetIpCountry(ipAddress string) (countryCode string, err error) {
	cachedResult := cache.Get(ipAddress)
	if cachedResult != "" {
		return cachedResult, nil
	}

	whoisInfo, err := ripego.IPLookup(ipAddress)
	if err != nil {
		return "unknown", fmt.Errorf("whois lookup error for '%s': %s \n", ipAddress, err.Error())
	}

	if whoisInfo.Country == "" {
		return "unknown", errors.New("whois results contain no country")
	}

	countryCode = strings.ToLower(nonAlphanumericRegex.ReplaceAllString(strings.ToLower(whoisInfo.Country), ""))

	cache.Save(countryCode, ipAddress)

	return countryCode, nil
}
