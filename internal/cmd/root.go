package cmd

import (
	"fmt"
	"os"

	trqhelper "github.com/Donders-Institute/hpc-torque-helper/pkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Verbose is the flag to switch on/off the verbosed output of commands.
var Verbose bool

// TorqueServerHost is the hostname of the Torque server.
var TorqueServerHost string

// TorqueHelperPort is the port number of the Torque Helper service.
var TorqueHelperPort int

var logger = logrus.New()

var xml bool

func init() {

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")

	qstatCmd.Flags().BoolVarP(&xml, "xml", "x", false, "XML output")
	qstatCmd.Flags().StringVarP(&TorqueServerHost, "server", "s", "torque.dccn.nl", "Torque server hostname")
	qstatCmd.Flags().IntVarP(&TorqueHelperPort, "port", "p", 60209, "Torque helper service port")

	configCmd.Flags().StringVarP(&TorqueServerHost, "server", "s", "torque.dccn.nl", "Torque server hostname")
	configCmd.Flags().IntVarP(&TorqueHelperPort, "port", "p", 60209, "Torque helper service port")

	rootCmd.AddCommand(initCmd, qstatCmd, configCmd)
}

var rootCmd = &cobra.Command{
	Use:   "cluster-tool",
	Short: "Unified CLI for various HPC cluster utilities.",
	Long:  `A unified command-line interface for different HPC cluster utilities.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize with the (sub-)command auto completion.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Create("cluster-tool")
		if err != nil {
			panic(fmt.Errorf("cannot open file: cluster"))
		}
		defer f.Close()
		rootCmd.GenBashCompletion(f)
	},
}

var qstatCmd = &cobra.Command{
	Use:   "qstat",
	Short: "Print job list in the memory of the Torque server.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		trqhelper.PrintClusterQstat(TorqueServerHost, TorqueHelperPort, xml)
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print Torque and Moab server configurations.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		trqhelper.PrintClusterConfig(TorqueServerHost, TorqueHelperPort)
	},
}

// Execute is the main entry point of the cluster command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
