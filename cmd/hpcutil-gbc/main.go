// This program generates the Bash completion script and prints it on the stdout.
package main

import (
	"os"

	"github.com/Donders-Institute/hpc-utility/internal/cmd"
)

func main() {
	hpcutil := cmd.NewHpcutilCmd()
	hpcutil.GenBashCompletion(os.Stdout)
}
