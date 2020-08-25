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
	"fmt"
	"os"
	"os/signal"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ks07/t11c-reset/pkg/t11c"
)

var (
	cfgFile  string
	verbose  bool
	dryrun   bool
	username string
	password string
	hostname string

	cancel context.CancelFunc
	ctx    context.Context
	conn   *t11c.Connection
	logger log.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "t11c-reset",
	Short: "A tool for resetting the modem on a Zyxel AMG1302-T11C",
	Long: `A tool for resetting the ADSL modem on a Zyxel AMG1302-T11C. The
tool interacts with the web interface, and must be provided the credentials
and hostname of the router.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		var levelLimit level.Option
		if viper.GetBool("verbose") {
			levelLimit = level.AllowAll()
		} else {
			levelLimit = level.AllowInfo()
		}
		logger = level.NewFilter(logger, levelLimit)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)

		baseCtx := context.Background()
		ctx, cancel = context.WithCancel(baseCtx)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		defer signal.Stop(c)

		go func() {
			select {
			case <-c:
				cancel()
			case <-ctx.Done():
				return
			}
		}()

		conn = t11c.NewConnection(logger, viper.GetBool("no-action"), viper.GetString("username"), viper.GetString("password"), viper.GetString("hostname"))
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		cancel()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		level.Error(logger).Log("msg", "failed to start", "err", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// These persistent flags are global across the application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.t11c-reset.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "write verbose logging output")

	rootCmd.PersistentFlags().BoolVarP(&dryrun, "no-action", "n", false, "Don't make changes to the modem")
	rootCmd.PersistentFlags().StringVar(&username, "username", "admin", "The username to login with")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "The password to login with")
	rootCmd.PersistentFlags().StringVar(&hostname, "hostname", "192.168.1.1", "The hostname or IP of the router")

	// Flags may be passed via environment variables with this prefix
	viper.SetEnvPrefix("T11C_")

	viper.BindPFlags(rootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".t11c-reset" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".t11c-reset")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
