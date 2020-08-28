package internal

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/ks07/t11c-reset/pkg/net"
	"github.com/ks07/t11c-reset/pkg/t11c"
)

func WatchReset(ctx context.Context, logger log.Logger, conn *t11c.Connection, interval uint, privileged bool, remoteHost string) {
	level.Info(logger).Log("interval", interval, "remote_host", remoteHost, "msg", "starting monitoring")

	checker := net.NewPingChecker(remoteHost, privileged)

	// Run a check immediately, unless the context has already been cancelled
	select {
	case <-ctx.Done():
		level.Info(logger).Log("msg", "monitoring cancelled")
		return
	default:
		checkReset(ctx, logger, conn, checker)
	}

	// After the initial check, start the ticker which will first trigger after the interval
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			level.Info(logger).Log("msg", "monitoring cancelled")
			return
		case <-ticker.C:
			checkReset(ctx, logger, conn, checker)
		}
	}
}

func checkReset(ctx context.Context, logger log.Logger, conn *t11c.Connection, checker net.PingChecker) {
	up, err := checker.CheckRemoteConnectivity(ctx, logger)
	if err != nil {
		level.Error(logger).Log("msg", "failed to start connectivity tests", "err", err)
		return
	}

	if up {
		level.Debug(logger).Log("msg", "connectivity ok")
		return
	}

	level.Info(logger).Log("msg", "connection is down")
	for {
		if err := resetAndWait(ctx, logger, conn, checker); err != nil {
			level.Warn(logger).Log("msg", "modem reset failed", "err", err)
		} else {
			level.Info(logger).Log("msg", "connection restored")
			break
		}
	}
}

func resetAndWait(ctx context.Context, logger log.Logger, conn *t11c.Connection, checker net.PingChecker) error {
	level.Info(logger).Log("msg", "resetting modem")

	valid, err := conn.TestSession(ctx)
	if err != nil {
		level.Error(logger).Log("msg", "failed to check session", "err", err)
		return err
	}

	if !valid {
		if err := conn.Login(ctx); err != nil {
			level.Error(logger).Log("msg", "failed to login", "err", err)
			return err
		}
	}

	if err := conn.SetModemState(ctx, false); err != nil {
		// If explicit disconnection fails, just attempt to connect anyway
		level.Warn(logger).Log("msg", "failed to disconnect, will try reconnect alone", "err", err)
	}

	if err := conn.SetModemState(ctx, true); err != nil {
		level.Error(logger).Log("msg", "failed to reconnect modem", "err", err)
		return err
	}

	level.Info(logger).Log("msg", "reset complete, waiting for connectivity")
	err = checker.WaitForRemoteConnectivity(ctx, logger)
	return err
}
