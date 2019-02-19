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
	fmt.Printf("\nTrace the job.\n")
	fmt.Printf("\nUSAGE: %s [OPTIONS] jobId\n", os.Args[0])
	fmt.Printf("\nOPTIONS:\n")
	flag.PrintDefaults()
	fmt.Printf("\n")
}

func main() {

	// command-line arguments
	args := flag.Args()

	if len(args) != 1 {
		flag.Usage()
		log.Fatal(fmt.Sprintf("require one job id: %+v", args))
	}

	cli.PrintClusterTracejob(args[0], *trqhelpdHost, *trqhelpdPort)
}
