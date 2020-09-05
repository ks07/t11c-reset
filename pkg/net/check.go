package net

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sparrc/go-ping"
)

const downThreshold = 100.0 // The packet loss proportion below which the connection is considered up
const pingsRequired = 2     // The number of responses required before the connection is deemed restored

type PingChecker struct {
	Dests     []string
	RawSocket bool
}

func NewPingChecker(dest []string, rawSocket bool) PingChecker {
	return PingChecker{
		Dests:     dest,
		RawSocket: rawSocket,
	}
}

func (pc PingChecker) makePinger(ctx context.Context, dest string) (*ping.Pinger, error) {
	pinger, err := ping.NewPinger(dest)
	if err != nil {
		return nil, err
	}

	// Run platform-specific setup for the ping socket
	platformSetupPinger(pinger, pc.RawSocket)

	// Immediately stop pinging if the context is cancelled
	go func() {
		<-ctx.Done()
		pinger.Stop()
	}()

	return pinger, nil
}

func (pc PingChecker) CheckRemoteConnectivity(ctx context.Context, logger log.Logger) (bool, error) {
	pingerCtx, pingerCancel := context.WithCancel(ctx)
	defer pingerCancel()

	for _, dest := range pc.Dests {
		pinger, err := pc.makePinger(pingerCtx, dest)
		if err != nil {
			return false, err
		}

		// We could just use pinger's interval setting, but we specifically want to run bursts in case of random packet loss
		pinger.Count = 3
		pinger.Timeout = 6 * time.Second

		pinger.Run()
		stats := pinger.Statistics()

		// Need to check pings were actually attempted, as a workaround for https://github.com/sparrc/go-ping/issues/92
		if stats.PacketsSent == 0 {
			return false, errors.New("no ping packets were sent, potential configuration error")
		}

		level.Debug(logger).Log("remote_host", dest, "packets_sent", stats.PacketsSent, "packets_dropped", stats.PacketsSent-stats.PacketsRecv, "latency", stats.AvgRtt, "msg", "ping complete")

		// Only try the next host if the test failed
		if stats.PacketLoss < downThreshold {
			return true, nil
		}
	}

	return false, nil
}

func (pc PingChecker) WaitForRemoteConnectivity(ctx context.Context, logger log.Logger) error {
	var wg sync.WaitGroup
	var pingsReceived uint32
	pingerCtx, pingerCancel := context.WithCancel(ctx)
	defer pingerCancel()

	var pingers []*ping.Pinger
	for _, dest := range pc.Dests {
		pinger, err := pc.makePinger(pingerCtx, dest)
		if err != nil {
			return err
		}

		pinger.Timeout = 30 * time.Second

		pinger.OnRecv = func(pkt *ping.Packet) {
			newPingsReceived := atomic.AddUint32(&pingsReceived, 1)
			level.Debug(logger).Log("remote_host", dest, "packets_received", newPingsReceived, "latency", pkt.Rtt, "msg", "connection restored")
			if newPingsReceived >= pingsRequired {
				pingerCancel()
			}
		}

		pingers = append(pingers, pinger)
	}

	for _, pinger := range pingers {
		wg.Add(1)
		go func() {
			pinger.Run()
			wg.Done()
		}()
	}

	wg.Wait()

	if pingsReceived < pingsRequired {
		return errors.New("connection did not come back up")
	}
	return nil
}
