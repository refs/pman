package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"

	"github.com/refs/pman/pkg/controller"
	"github.com/refs/pman/pkg/process"
	"github.com/refs/pman/pkg/service"
)

var (
	extRun  = flag.String("run", "", "oCIS extension to run")
	extKill = flag.String("kill", "", "oCIS extension to terminate")
	l       = flag.Bool("l", false, "list running extensions")
)

func main() {
	flag.Parse()
	if *l {
		client, err := rpc.DialHTTP("tcp", "localhost:10666")
		if err != nil {
			log.Fatal("dialing:", err)
		}

		var arg1 string

		if err := client.Call("Service.List", struct{}{}, &arg1); err != nil {
			log.Fatal(err)
		}

		fmt.Println(arg1)
		return
	}
	if *extKill != "" {
		client, err := rpc.DialHTTP("tcp", "localhost:10666")
		if err != nil {
			log.Fatal("dialing:", err)
		}

		var arg1 int

		if err := client.Call("Service.Kill", &*extKill, &arg1); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *extRun != "" {
		client, err := rpc.DialHTTP("tcp", "localhost:10666")
		if err != nil {
			log.Fatal("dialing:", err)
		}

		arg0 := process.NewProcEntry(
			*extRun,
			[]string{*extRun}...,
		)
		var arg1 int

		if err := client.Call("Service.Start", arg0, &arg1); err != nil {
			log.Fatal(err)
		}
	} else {
		service.Start(
			controller.WithBinary("ocis"), // Use case: binary rename.
		)
	}
}
