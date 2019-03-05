package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Verbose is the flag to switch on/off the verbosed output of commands.
var Verbose bool

// TorqueServerHost is the hostname of the Torque server.
var TorqueServerHost string

// TorqueHelperPort is the port number of the Torque Helper service.
var TorqueHelperPort int

// TorqueHelperCert is the path of the TorqueHelper server certificate.
var TorqueHelperCert string

// NetDomain is the default network domain name.
// It allows commands to accept short hostname specification in arguments.
var NetDomain string

// NewHpcutilCmd returns the root command.
func NewHpcutilCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&NetDomain, "domain", "d", "dccn.nl", "default network domain")
}

var rootCmd = &cobra.Command{
	Use:   "hpcutil",
	Short: "Unified CLI for various HPC cluster utilities.",
	Long:  `A unified command-line interface for different HPC cluster utilities.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("verbose") {
			log.SetLevel(log.DebugLevel)
		}
	},
}

// Execute is the main entry point of the cluster command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
}
