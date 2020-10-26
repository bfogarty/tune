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
  tune to --localPort=<port> <host> <port>
  tune -h | --help
  tune --version

Options:
  -h --help           Show this screen.
  --version           Show version.
  --localPort=<port>  The local port to use.`

        // parse arguments
        opts, err := docopt.ParseArgs(usage, os.Args[1:], "0.1.0")
        checkErr(err)

        remoteHost, err := opts.String("<host>")
        checkErr(err)

        remotePort, err := opts.Int("<port>")
        checkErr(err)

        localPort, err := opts.Int("--localPort")
        checkErr(err)

        // create and start a new tunnel
        t, err := tunnel.New(remoteHost, localPort, remotePort)
        checkErr(err)

        err = t.Start()
        checkErr(err)
}

func checkErr(err error) {
        if err != nil {
            log.Fatal(err)
        }
}
