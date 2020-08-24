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
	"context"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks the reported connection state of the modem",
	Long: `Checks the connection state of the modem as reported
by the UI. This may not always reflect the actual connection
state, as there is some delay before the modem detects a drop.

Exits with a code of 2 if the modem reports as disconnected.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if err := conn.Login(ctx); err != nil {
			level.Error(logger).Log("msg", "failed to login", "err", err)
			return
		}

		connected, err := conn.ModemIsConnected(ctx)
		if err != nil {
			level.Error(logger).Log("msg", "failed to check state", "err", err)
			return
		}

		if connected {
			level.Info(logger).Log("msg", "Modem believes connection is active")
		} else {
			level.Info(logger).Log("msg", "Modem believes connection is down")
			os.Exit(2)
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
