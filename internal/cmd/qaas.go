package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var qaasHost string
var qaasPort int
var webhookName string

func init() {

	webhookCmd.PersistentFlags().StringVarP(&qaasHost, "server", "s", "qaas.dccn.nl", "QaaS service hostname")
	webhookCmd.PersistentFlags().IntVarP(&qaasPort, "port", "p", 5111, "QaaS service hostname")

	createCmd.Flags().StringVarP(&webhookName, "name", "n", "MyHook", "name of the webhook")

	webhookCmd.AddCommand(createCmd, deleteCmd, infoCmd, triggerCmd)
	qaasCmd.AddCommand(webhookCmd)
	rootCmd.AddCommand(qaasCmd)
}

var qaasCmd = &cobra.Command{
	Use:   "qaas",
	Short: "Perform an action on the Qsub-as-a-Service (QaaS).",
	Long:  ``,
}

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage webhooks.",
	Long:  ``,
}

var createCmd = &cobra.Command{
	Use:   "create [ScriptPath]",
	Short: "Create a new webhook.",
	Long:  ``,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expect 1 script, given %d", len(args))
		}
		// check if args[0] is a valid file
		s, err := os.Stat(args[0])
		if err != nil {
			return err
		}
		if !s.Mode().IsRegular() {
			return fmt.Errorf("not a regular file: %s", args[0])
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Warnf("Not implemented!!")
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete an existing webhook.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger.Warnf("Not implemented!!")
	},
}

var infoCmd = &cobra.Command{
	Use:   "info [name]",
	Short: "Retrieve information of an existing webhook.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger.Warnf("Not implemented!!")
	},
}

var triggerCmd = &cobra.Command{
	Use:   "trigger [name]",
	Short: "Trigger an existing webhook manually.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger.Warnf("Not implemented!!")
	},
}
