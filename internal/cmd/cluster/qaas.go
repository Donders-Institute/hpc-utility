package cluster

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(qaasCmd)
}

var qaasCmd = &cobra.Command{
	Use:   "qaas",
	Short: "Perform an action on the Qsub-as-a-Service (QaaS).",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Print: " + strings.Join(args, " "))
	},
}
