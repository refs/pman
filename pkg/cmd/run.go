package cmd

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"

	"github.com/refs/pman/pkg/config"
	"github.com/refs/pman/pkg/process"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Run an extension.
func Run(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run an extension.",
		Args:  cobra.MinimumNArgs(1),
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

			proc := process.NewProcEntry(args[0], os.Environ(), []string{args[0]}...)
			var res int

			if err := client.Call("Service.Start", proc, &res); err != nil {
				log.Fatal(err)
			}

			fmt.Println(res)
		},
	}
}
