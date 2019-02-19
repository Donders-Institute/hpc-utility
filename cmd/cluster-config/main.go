package main

import (
	"flag"
	"fmt"
	"os"

	cli "github.com/Donders-Institute/hpc-torque-helper/pkg"

	log "github.com/sirupsen/logrus"
)

var (
	trqhelpdHost *string
	trqhelpdPort *int
	optsVerbose  *bool
)

func init() {

	// Command-line arguments
	trqhelpdHost = flag.String("h", "torque.dccn.nl", "set the service `host` of the trqhelpd")
	trqhelpdPort = flag.Int("p", 60209, "set the service `port` of the trqhelpd")
	optsVerbose = flag.Bool("v", false, "print debug messages")
	flag.Usage = usage

	flag.Parse()

	// set logging
	log.SetOutput(os.Stderr)
	// set logging level
	llevel := log.InfoLevel
	if *optsVerbose {
		llevel = log.DebugLevel
	}
	log.SetLevel(llevel)
	// set logging
	log.SetOutput(os.Stderr)
}

func usage() {
	fmt.Printf("\nGet the Torque and Moab configuration.\n")
	fmt.Printf("\nUSAGE: %s [OPTIONS]\n", os.Args[0])
	fmt.Printf("\nOPTIONS:\n")
	flag.PrintDefaults()
	fmt.Printf("\n")
}

func main() {
	cli.PrintClusterConfig(*trqhelpdHost, *trqhelpdPort)
}
