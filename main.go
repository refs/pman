package main

import (
	"log"

	"github.com/refs/pman/pkg/cmd"
	"github.com/refs/pman/pkg/config"
)

func main() {
	if err := cmd.RootCmd(config.NewConfig()).Execute(); err != nil {
		log.Fatal(err)
	}
}
