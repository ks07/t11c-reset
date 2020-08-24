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
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			level.Info(logger).Log("msg", "monitoring cancelled")
			return
		case <-ticker.C:
			checkReset(ctx, logger, conn, privileged, remoteHost)
		}
	}
}

func checkReset(ctx context.Context, logger log.Logger, conn *t11c.Connection, privileged bool, remoteHost string) {
	up, err := net.CheckRemoteConnectivity(ctx, logger, remoteHost, privileged)
	if err != nil {
		level.Error(logger).Log("msg", "failed to start connectivity tests", "err", err)
		return
	}

	if up {
		level.Debug(logger).Log("msg", "connectivity ok")
		return
	}

	level.Info(logger).Log("msg", "connection is down")
	valid, err := conn.TestSession(ctx)
	if err != nil {
		level.Error(logger).Log("msg", "failed to check session", "err", err)
		return
	}

	if !valid {
		if err := conn.Login(ctx); err != nil {
			level.Error(logger).Log("msg", "failed to login", "err", err)
			return
		}
	}

	if err := conn.SetModemState(ctx, false); err != nil {
		// If explicit disconnection fails, just attempt to connect anyway
		level.Warn(logger).Log("msg", "failed to disconnect, will try reconnect alone", "err", err)
	}

	if err := conn.SetModemState(ctx, true); err != nil {
		level.Error(logger).Log("msg", "failed to reconnect modem", "err", err)
		return
	}
}
