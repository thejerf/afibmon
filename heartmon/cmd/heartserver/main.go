package main

import (
	"flag"
	"fmt"

	"github.com/thejerf/afibmon/heartmon"
	"github.com/thejerf/suture"
)

var address = flag.String("address", ":18498", "the address to bind the server to")

func main() {
	flag.Parse()

	supervisor := suture.NewSimple("heartmon supervisor")

	server, err := heartmon.NewServer(*address)
	if err != nil {
		panic("Can't bind: " + err.Error())
	}
	supervisor.Add(server)

	fmt.Println("Beginning serving")

	supervisor.Serve()
}
