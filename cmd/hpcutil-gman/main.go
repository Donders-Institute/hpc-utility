// This program generates the Linux online manul to /tmp/hpcutil.1
package main

import (
	"log"

	"github.com/Donders-Institute/hpc-utility/internal/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	hpcutil := cmd.NewHpcutilCmd()

	header := &doc.GenManHeader{
		Title:   "hpcutil",
		Section: "1",
	}
	err := doc.GenManTree(hpcutil, header, "/tmp")
	if err != nil {
		log.Fatal(err)
	}
}
