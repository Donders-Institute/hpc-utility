package cmd

import (
	trqhelper "github.com/Donders-Institute/hpc-torque-helper/pkg/client"
	"github.com/spf13/cobra"
)

func init() {
	jobCmd.AddCommand(jobTraceCmd, jobMeminfoCmd)
	jobCmd.PersistentFlags().StringVarP(&TorqueServerHost, "server", "s", "torque.dccn.nl", "Torque server hostname")
	jobCmd.PersistentFlags().IntVarP(&TorqueHelperPort, "port", "p", 60209, "Torque helper service port")
	jobCmd.PersistentFlags().StringVarP(&TorqueHelperCert, "cert", "c", "", "Torque helper service certificate")

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
		c := trqhelper.TorqueHelperSrvClient{
			SrvHost:     TorqueServerHost,
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		c.PrintClusterTracejob(args[0])
	},
}

var jobMeminfoCmd = &cobra.Command{
	Use:   "meminfo [JobID]",
	Short: "Print memory usage of a running job.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := trqhelper.TorqueHelperMomClient{
			SrvHost:     TorqueServerHost,
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		c.PrintJobMemoryInfo(args[0])
	},
}
