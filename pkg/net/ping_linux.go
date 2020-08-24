package net

import "github.com/sparrc/go-ping"

// On Linux platforms, we need to ask the user, as unprivileged ping may be disabled, and raw sockets require a capab
func platformSetupPinger(pinger *ping.Pinger, rawSocket bool) {
	pinger.SetPrivileged(rawSocket)
}
