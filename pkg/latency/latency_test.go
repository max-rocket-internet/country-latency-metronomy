package latency

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLatency(t *testing.T) {
	destination := Destination{Ip: "aaaa"}
	result, err := GetLatency(destination)
	assert.False(t, result.Successful)
	assert.ErrorContains(t, err, "unable to parse IP")

	destination = Destination{Ip: ""}
	result, err = GetLatency(destination)
	assert.False(t, result.Successful)
	assert.ErrorContains(t, err, "Destination IP is blank")

	destination = Destination{Ip: "192.168.0.1"}
	result, err = GetLatency(destination)
	assert.False(t, result.Successful)
	assert.ErrorContains(t, err, "Private IP address (RFC 1918) given '192.168.0.1'")

	destination = Destination{Ip: "8.8.8.8", CountryCode: "aaa"}
	result, err = GetLatency(destination)
	assert.False(t, result.Successful)
	assert.ErrorContains(t, err, "Country code should be 2 characters (ISO 3166-1 alpha-2)")

	// This test needs root priviliges
	// destination = Destination{Ip: "8.8.8.8"}
	// result, err = GetLatency(destination)
	// assert.True(t, result.Successful)
	// assert.Zero(t, result.Latency.PacketLoss)
	// assert.Equal(t, result.AlternativeIP, "")
	// assert.NoError(t, err)

	// This test needs root priviliges
	// destination = Destination{Ip: "82.88.43.13"}
	// result, err = GetLatency(destination)
	// assert.True(t, result.Successful)
	// assert.NoError(t, err)
	// assert.NotEmpty(t, result.AlternativeIP)
	// assert.Equal(t, result.CountryCode, "it")
}
