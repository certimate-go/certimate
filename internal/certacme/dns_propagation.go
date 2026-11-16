package certacme

import (
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"

	"github.com/certimate-go/certimate/pkg/core"
)

func withDNSPropagationTiming(provider core.ACMEChallenger, timeoutSeconds, pollingSeconds int) core.ACMEChallenger {
	if timeoutSeconds <= 0 && pollingSeconds <= 0 {
		return provider
	}

	challenger := dnsChallengerWithTiming{
		ACMEChallenger: provider,
		timeout:        time.Duration(timeoutSeconds) * time.Second,
		interval:       time.Duration(pollingSeconds) * time.Second,
	}
	if sequential, ok := provider.(interface{ Sequential() time.Duration }); ok {
		return sequentialDNSChallengerWithTiming{challenger, sequential.Sequential}
	}
	return challenger
}

type dnsChallengerWithTiming struct {
	core.ACMEChallenger
	timeout  time.Duration
	interval time.Duration
}

func (c dnsChallengerWithTiming) Timeout() (time.Duration, time.Duration) {
	timeout, interval := dns01.DefaultPropagationTimeout, dns01.DefaultPollingInterval
	if timing, ok := c.ACMEChallenger.(challenge.ProviderTimeout); ok {
		timeout, interval = timing.Timeout()
	}
	if c.timeout > 0 {
		timeout = c.timeout
	}
	if c.interval > 0 {
		interval = c.interval
	}
	return timeout, interval
}

type sequentialDNSChallengerWithTiming struct {
	dnsChallengerWithTiming
	sequenceInterval func() time.Duration
}

func (c sequentialDNSChallengerWithTiming) Sequential() time.Duration {
	return c.sequenceInterval()
}
