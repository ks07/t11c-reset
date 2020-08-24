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
	"fmt"

	"github.com/spf13/cobra"
)

var connectOnly bool

// reconnectCmd represents the reconnect command
var reconnectCmd = &cobra.Command{
	Use:   "reconnect",
	Short: "Immediately disconnects and reconnects the ADSL connection of the modem.",
	Long: `This command will immediately disconnect and reconnect the ADSL connection
of the modem. This is useful if the connection has dropped but the modem has not yet
detected the failure state.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := conn.Login(); err != nil {
			fmt.Println(err)
			return
		}

		ok, err := conn.TestSession()
		if err != nil {
			fmt.Println(err)
			return
		}
		if !ok {
			fmt.Println("Login failed")
			return
		}
		fmt.Println("Login complete, resetting connection...")

		if !connectOnly {
			if err = conn.SetModemState(false); err != nil {
				fmt.Println("Disconnect failed")
			}
		}

		if err = conn.SetModemState(true); err != nil {
			fmt.Println("Reconnect failed")
			return
		}

		fmt.Println("Done!")
	},
}

func init() {
	rootCmd.AddCommand(reconnectCmd)

	reconnectCmd.Flags().BoolVarP(&connectOnly, "connect-only", "c", false, "Skip the disconnect step")
}
