package cmd

import (
	dg "github.com/Donders-Institute/hpc-cluster-tools/internal/datagetter"
	trqhelper "github.com/Donders-Institute/hpc-torque-helper/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var xml bool

const (
	gib float64 = 1024 * 1024 * 1024
)

// variable may be set at the build time to fix the default location for the TorqueHelper server certificate.
var defTorqueHelperCert string

func init() {

	qstatCmd.Flags().BoolVarP(&xml, "xml", "x", false, "XML output")

	clusterCmd.PersistentFlags().StringVarP(&TorqueServerHost, "server", "s", "torque.dccn.nl", "Torque server hostname")
	clusterCmd.PersistentFlags().IntVarP(&TorqueHelperPort, "port", "p", 60209, "Torque helper service port")
	clusterCmd.PersistentFlags().StringVarP(&TorqueHelperCert, "cert", "c", defTorqueHelperCert, "Torque helper service certificate")

	nodeCmd.AddCommand(nodeMeminfoCmd, nodeDiskinfoCmd, nodeVncCmd)
	jobCmd.AddCommand(jobTraceCmd, jobMeminfoCmd)
	clusterCmd.AddCommand(qstatCmd, configCmd, jobCmd, nodeCmd)

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

// job related subcommands
var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Retrieve information about a cluster job.",
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
		if err := c.PrintClusterTracejob(args[0]); err != nil {
			log.Errorf("fail get job trace info: %+v\n", err)
		}
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
		if err := c.PrintJobMemoryInfo(args[0]); err != nil {
			log.Errorf("fail get job memory utilisation: %+v\n", err)
		}
	},
}

// node related subcommands
type nodeType uint

const (
	access nodeType = iota
	compute
)

var nodeTypeNames = map[nodeType]string{
	access:  "access",
	compute: "compute",
}

var nodeCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Retrieve information about cluster nodes.",
	Long:  ``,
	// ValidArgs: []string{nodeTypeNames[access], nodeTypeNames[compute]},
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// TODO: get nodes overview
	// },
}

var nodeMeminfoCmd = &cobra.Command{
	Use:       "memfree",
	Short:     "Print total and free memory of the cluster's access nodes.",
	Long:      ``,
	ValidArgs: []string{nodeTypeNames[access], nodeTypeNames[compute]},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = []string{nodeTypeNames[access], nodeTypeNames[compute]}
		}
		for _, n := range args {
			switch n {
			case nodeTypeNames[access]:
				g := dg.GangliaDataGetter{Dataset: dg.MemoryUsageAccessNode}
				g.GetPrint()
			case nodeTypeNames[compute]:
				g := dg.GangliaDataGetter{Dataset: dg.MemoryUsageComputeNode}
				g.GetPrint()
			}
		}
	},
}

var nodeDiskinfoCmd = &cobra.Command{
	Use:       "diskfree",
	Short:     "Print total and free disk space of the cluster's access nodes.",
	Long:      ``,
	ValidArgs: []string{nodeTypeNames[access], nodeTypeNames[compute]},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = []string{nodeTypeNames[access], nodeTypeNames[compute]}
		}
		for _, n := range args {
			switch n {
			case nodeTypeNames[access]:
				g := dg.GangliaDataGetter{Dataset: dg.DiskUsageAccessNode}
				g.GetPrint()
			case nodeTypeNames[compute]:
				g := dg.GangliaDataGetter{Dataset: dg.DiskUsageComputeNode}
				g.GetPrint()
			}
		}
	},
}

var nodeVncCmd = &cobra.Command{
	Use:   "vnc",
	Short: "List vnc servers running on a given node.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := trqhelper.TorqueHelperAccClient{
			SrvHost:     args[0],
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		servers, err := c.GetVNCServers()
		if err != nil {
			log.Errorln(err)
			return
		}
		for _, s := range servers {
			log.Infof("%s %s\n", s.Owner, s.ID)
		}
	},
}
