package net

import "github.com/sparrc/go-ping"

// On windows platforms, we can always use raw ICMP
func platformSetupPinger(pinger *ping.Pinger, _ bool) {
	pinger.SetPrivileged(true)
}
