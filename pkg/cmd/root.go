package cmd

import (
	"github.com/refs/pman/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "pman",
		Short: "RPC Process Manager",
	}

	Hostname string
	Port     string
	KeepAlive  bool
)

// RootCmd returns a configured root command.
func RootCmd(cfg *config.Config) *cobra.Command {
	rootCmd.PersistentFlags().StringVarP(&Hostname, "hostname", "n", "", "host with a running ocis runtime.")
	rootCmd.PersistentFlags().StringVarP(&Port, "port", "p", "", "port to send messages to the rpc ocis runtime.")
	rootCmd.PersistentFlags().BoolVarP(&KeepAlive, "keep-alive", "k", false, "restart supervised processes that abruptly die.")

	viper.BindPFlag("hostname", rootCmd.PersistentFlags().Lookup("hostname"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("keep-alive", rootCmd.PersistentFlags().Lookup("keep-alive"))

	viper.AutomaticEnv()
	viper.BindEnv("keep-alive", "RUNTIME_KEEP_ALIVE")

	rootCmd.AddCommand(List(cfg))
	rootCmd.AddCommand(Run(cfg))
	rootCmd.AddCommand(Kill(cfg))

	return rootCmd
}
