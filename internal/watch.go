package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/ks07/t11c-reset/pkg/net"
	"github.com/ks07/t11c-reset/pkg/t11c"
)

func WatchReset(ctx context.Context, conn *t11c.Connection, interval uint, privileged bool, remoteHost string) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkReset(ctx, conn, privileged, remoteHost)
		}
	}
}

func checkReset(ctx context.Context, conn *t11c.Connection, privileged bool, remoteHost string) {
	up, err := net.CheckRemoteConnectivity(ctx, remoteHost, privileged)
	if err != nil {
		fmt.Println(err)
		return
	}

	if up {
		fmt.Println("Connection is alive")
		return
	}

	fmt.Println("Connection is down, reconnecting...")
	valid, err := conn.TestSession()
	if err != nil {
		fmt.Println(err)
		return
	}

	if !valid {
		if err := conn.Login(); err != nil {
			fmt.Println(err)
			return
		}
	}

	if err := conn.SetModemState(false); err != nil {
		// If explicit disconnection fails, just attempt to connect anyway
		fmt.Println(err)
	}

	if err := conn.SetModemState(true); err != nil {
		fmt.Println(err)
		return
	}
}
