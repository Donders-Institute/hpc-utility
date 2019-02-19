package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	cli "github.com/Donders-Institute/hpc-torque-helper/pkg"
)

var (
	trqHelpdPort *int
	optsVerbose  *bool
)

func init() {

	// Command-line arguments
	trqHelpdPort = flag.Int("p", 60209, "set the `port` number of the trqhelpd")
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
	fmt.Printf("\nGet the memory usage information of a job.\n")
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

	cli.PrintJobMemoryInfo(args[0], *trqHelpdPort)
}
