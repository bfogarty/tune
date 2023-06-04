package main

import (
	"github.com/bfogarty/tune/pkg/tunnel"
	"github.com/docopt/docopt-go"
	"log"
	"os"
)

func main() {
	usage := `tune

Usage:
  tune to --localPort=<port> [--region=<region>] <host> <port>
  tune -h | --help
  tune --version

Options:
  -h --help           Show this screen.
  --version           Show version.
  --region=<region>   The AWS region to use [default: us-east-1].
  --localPort=<port>  The local port to use.`

	// parse arguments
	opts, err := docopt.ParseArgs(usage, os.Args[1:], "0.2.0")
	checkErr(err)

	remoteHost, err := opts.String("<host>")
	checkErr(err)

	remotePort, err := opts.Int("<port>")
	checkErr(err)

	localPort, err := opts.Int("--localPort")
	checkErr(err)

	region, err := opts.String("--region")
	checkErr(err)

	// create and start a new tunnel
	t, err := tunnel.New(remoteHost, localPort, remotePort, region)
	checkErr(err)

	err = t.Start()
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
