package net

import (
	"context"
	"time"

	"github.com/sparrc/go-ping"
)

const downThreshold = 1.0

func CheckRemoteConnectivity(ctx context.Context, dest string, rawSocket bool) (bool, error) {
	pinger, err := ping.NewPinger(dest)
	if err != nil {
		return false, err
	}

	// We could just use pinger's interval setting, but we specifically want to run bursts in case of random packet loss
	pinger.Count = 3
	pinger.Timeout = 6 * time.Second

	// Run platform-specific setup for the ping socket
	platformSetupPinger(pinger, rawSocket)

	// Immediately stop pinging if the context is cancelled
	go func() {
		for _ = range ctx.Done() {
			pinger.Stop()
		}
	}()

	pinger.Run()

	return pinger.Statistics().PacketLoss < downThreshold, nil
}
