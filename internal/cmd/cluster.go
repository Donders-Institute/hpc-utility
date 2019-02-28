package cmd

import (
	trqhelper "github.com/Donders-Institute/hpc-torque-helper/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var xml bool

// variable may be set at the build time to fix the default location for the TorqueHelper server certificate.
var defTorqueHelperCert string

func init() {

	qstatCmd.Flags().BoolVarP(&xml, "xml", "x", false, "XML output")

	clusterCmd.AddCommand(qstatCmd, configCmd, jobTraceCmd, jobMeminfoCmd)
	clusterCmd.PersistentFlags().StringVarP(&TorqueServerHost, "server", "s", "torque.dccn.nl", "Torque server hostname")
	clusterCmd.PersistentFlags().IntVarP(&TorqueHelperPort, "port", "p", 60209, "Torque helper service port")
	clusterCmd.PersistentFlags().StringVarP(&TorqueHelperCert, "cert", "c", defTorqueHelperCert, "Torque helper service certificate")

	rootCmd.AddCommand(clusterCmd)
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Retrieve information about the HPC cluster or a job.",
	Long:  ``,
}

var qstatCmd = &cobra.Command{
	Use:   "qstat",
	Short: "Print job list in the memory of the Torque server.",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("qstat command is triggerd.")
		c := trqhelper.TorqueHelperSrvClient{
			SrvHost:     TorqueServerHost,
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		if err := c.PrintClusterQstat(xml); err != nil {
			log.Errorf("%+v\n", err)
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print Torque and Moab server configurations.",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		if cmd.Flags().Changed("verbose") {
			log.SetLevel(log.DebugLevel)
		}

		c := trqhelper.TorqueHelperSrvClient{
			SrvHost:     TorqueServerHost,
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		if err := c.PrintClusterConfig(); err != nil {
			log.Errorf("%+v\n", err)
		}
	},
}

var jobTraceCmd = &cobra.Command{
	Use:   "jobtrace [JobID]",
	Short: "Print job's trace log available on the Torque server.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := trqhelper.TorqueHelperSrvClient{
			SrvHost:     TorqueServerHost,
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		if err := c.PrintClusterTracejob(args[0]); err != nil {
			log.Errorf("fail get job trace info: %+v\n", err)
		}
	},
}

var jobMeminfoCmd = &cobra.Command{
	Use:   "jobmeminfo [JobID]",
	Short: "Print memory usage of a running job.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := trqhelper.TorqueHelperMomClient{
			SrvHost:     TorqueServerHost,
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		if err := c.PrintJobMemoryInfo(args[0]); err != nil {
			log.Errorf("fail get job memory utilisation: %+v\n", err)
		}
	},
}
