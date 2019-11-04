package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// version is the CLI version string.
// It is a placeholder with value to be defined (with -X flag) during the code ompilation.
var defVersion string

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

// customized bash completion function for searching for webhook ids in the user's home directory.
const (
	funcBashCompletion = `__hpcutil_get_webhook_ids()
{
	local hpcutil_webhook_ids out
    if hpcutil_webhook_ids=$(hpcutil webhook list | egrep -e '^\S+' 2>/dev/null); then
        out=( $(echo "${hpcutil_webhook_ids}") )
        COMPREPLY=( $( compgen -W "${out[*]}" -- "$cur" ) )
	fi
	return 0
}

__custom_func() {
	case ${last_command} in 
		hpcutil_webhook_info | hpcutil_webhook_delete | hpcutil_webhook_trigger )
			__hpcutil_get_webhook_ids
			return
			;;
		*)
			;;
	esac
}
`
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&NetDomain, "domain", "d", "dccn.nl", "default network domain")
	rootCmd.AddCommand(versionCmd, availCmd)
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
	BashCompletionFunction: funcBashCompletion,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print CLI version.",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("HPC utility version: %s\n", defVersion)
	},
}

var availCmd = &cobra.Command{
	Use: "avail",
	Short: "Print availab sub-commands.",
	Long: ``,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetRowLine(true)
		// TODO: SetColMinWidth doesn't automatically wrap the long string accordingly.
		//table.SetColMinWidth(1,40)
		table.SetHeader([]string{"Command", "Description"})
		table.SetRowSeparator("-")
		addCommandUseToTable(rootCmd, "", table)
		table.Render()
	},
}

// addCommandUseToTable addes recursively command's and sub-commands' `Use` and `Short` to a 
// `tablewritter.Table`.
func addCommandUseToTable(cmd *cobra.Command, parentUse string, table *tablewriter.Table) {
	for _, c := range cmd.Commands() {

		// only prints the command itself if there is no further sub-commands
		if len(c.Commands()) == 0 {
			table.Append([]string{
				fmt.Sprintf("%s %s %s", parentUse, cmd.Use, c.Use),
				c.Short,
			})
		}
		addCommandUseToTable(c, fmt.Sprintf("%s %s", parentUse, cmd.Use), table)
	}
}

// Execute is the main entry point of the cluster command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
}
