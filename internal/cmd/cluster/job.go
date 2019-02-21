package cluster

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	trqhelper "github.com/Donders-Institute/hpc-torque-helper/pkg"
)

func init() {
	jobCmd.AddCommand(jobTraceCmd, jobMeminfoCmd)
	rootCmd.AddCommand(jobCmd)
}

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Retrieve privileged information of a Torque job.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Print: " + strings.Join(args, " "))
	},
}

var jobTraceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Print job's trace log available on the Torque server.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		trqhelper.PrintClusterTracejob(args[0], TorqueServerHost, TorqueHelperPort)
	},
}

var jobMeminfoCmd = &cobra.Command{
	Use:   "meminfo",
	Short: "Print memory usage of a running job.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		trqhelper.PrintJobMemoryInfo(args[0], TorqueHelperPort)
	},
}
