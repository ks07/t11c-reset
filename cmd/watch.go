/*
Copyright © 2020 George Field <george@cucurbit.dev>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ks07/t11c-reset/internal"
)

var (
	interval    uint
	privileged  bool
	remoteHosts []string
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Periodically checks external connectivity, forcing a reconnect if necessary",
	Long: `Regularly pings an external server to check for connectivity issues. If packets
are lost, then connect to the router and reset the modem.

If multiple remote hosts are specified, the additional hosts will be used to confirm a loss
of connectivity. In the event that the ping test to the first host fails, the second host
will be tested (and so on), and the connection will only be treated as down if all hosts fail.
This is useful to defend against outages on the remote end from triggering a reset.`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.WatchReset(ctx, logger, conn, interval, privileged, remoteHosts)
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().BoolVarP(&privileged, "raw-ping", "p", false, "Attempt to use raw sockets to send ping (ignored on Windows)")
	watchCmd.Flags().UintVarP(&interval, "interval", "i", 15, "The interval, in seconds, between ping tests")
	watchCmd.Flags().StringSliceVarP(&remoteHosts, "remote", "r", []string{"1.1.1.1"}, "The remote address to ping to test connectivity. May be specified multiple times to defend against remote outages.")
}
