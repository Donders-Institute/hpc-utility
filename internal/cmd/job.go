package cmd

import (
	"github.com/spf13/cobra"

	trqhelper "github.com/Donders-Institute/hpc-torque-helper/pkg"
)

func init() {
	jobCmd.AddCommand(jobTraceCmd, jobMeminfoCmd)
	jobCmd.PersistentFlags().StringVarP(&TorqueServerHost, "server", "s", "torque.dccn.nl", "Torque server hostname")
	jobCmd.PersistentFlags().IntVarP(&TorqueHelperPort, "port", "p", 60209, "Torque helper service port")

	rootCmd.AddCommand(jobCmd)
}

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Retrieve privileged information of a Torque job.",
	Long:  ``,
}

var jobTraceCmd = &cobra.Command{
	Use:   "trace [JobID]",
	Short: "Print job's trace log available on the Torque server.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		trqhelper.PrintClusterTracejob(args[0], TorqueServerHost, TorqueHelperPort)
	},
}

var jobMeminfoCmd = &cobra.Command{
	Use:   "meminfo [JobID]",
	Short: "Print memory usage of a running job.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		trqhelper.PrintJobMemoryInfo(args[0], TorqueHelperPort)
	},
}
