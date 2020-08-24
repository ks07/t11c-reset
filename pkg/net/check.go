package net

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sparrc/go-ping"
)

const downThreshold = 1.0

func CheckRemoteConnectivity(ctx context.Context, logger log.Logger, dest string, rawSocket bool) (bool, error) {
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
	stats := pinger.Statistics()

	// Need to check pings were actually attempted, as a workaround for https://github.com/sparrc/go-ping/issues/92
	if stats.PacketsSent == 0 {
		return false, errors.New("no ping packets were sent, potential configuration error")
	}

	level.Debug(logger).Log("packets_sent", stats.PacketsSent, "packets_dropped", stats.PacketsSent-stats.PacketsRecv, "latency", stats.AvgRtt, "msg", "ping complete")

	return stats.PacketLoss < downThreshold, nil
}
