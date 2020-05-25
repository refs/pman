package cmd

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/refs/pman/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// List running extensions.
func List(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"r"},
		Short:   "List running extensions",
		PreRun: func(cmd *cobra.Command, args []string) {
			hostname := viper.GetString("hostname")
			if hostname != "" {
				cfg.Hostname = hostname
			}

			port := viper.GetString("port")
			if port != "" {
				cfg.Port = port
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			client, err := rpc.DialHTTP("tcp", net.JoinHostPort(cfg.Hostname, cfg.Port))
			if err != nil {
				log.Fatal("dialing:", err)
			}

			var arg1 string

			if err := client.Call("Service.List", struct{}{}, &arg1); err != nil {
				log.Fatal(err)
			}

			fmt.Println(arg1)
		},
	}
}
