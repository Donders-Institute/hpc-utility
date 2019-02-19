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
	optsXML      *bool
	optsVerbose  *bool
)

func init() {

	// Command-line arguments
	trqhelpdHost = flag.String("h", "torque.dccn.nl", "set the service `host` of the trqhelpd")
	trqhelpdPort = flag.Int("p", 60209, "set the service `port` of the trqhelpd")
	optsXML = flag.Bool("x", false, "print jobs in XML format")
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
	fmt.Printf("\nList jobs of the HPC cluster.\n")
	fmt.Printf("\nUSAGE: %s [OPTIONS]\n", os.Args[0])
	fmt.Printf("\nOPTIONS:\n")
	flag.PrintDefaults()
	fmt.Printf("\n")
}

func main() {
	cli.PrintClusterQstat(*trqhelpdHost, *trqhelpdPort, *optsXML)
}
