package cmd

import (
	"fmt"

	qaas "github.com/Donders-Institute/hpc-qaas/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var qaasHost string
var qaasPort int
var qaasCertFile string
var webhookName string

// variable may be set at the build time to fix the default location for the QaaS server certificate.
var defQaasCert string

func init() {

	qaasCmd.PersistentFlags().StringVarP(&qaasHost, "server", "s", "qaas.dccn.nl", "QaaS service hostname")
	qaasCmd.PersistentFlags().IntVarP(&qaasPort, "port", "p", 5111, "QaaS service hostname")
	qaasCmd.PersistentFlags().StringVarP(&qaasCertFile, "cert", "c", defQaasCert, "QaaS service SSL certificate")

	createCmd.Flags().StringVarP(&webhookName, "name", "n", "MyHook", "name of the webhook")

	qaasCmd.AddCommand(createCmd, deleteCmd, infoCmd, triggerCmd, listCmd)
	rootCmd.AddCommand(qaasCmd)
}

var qaasCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage webhooks on the Qsub-as-a-Service (QaaS).",
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
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		webhook := qaas.Webhook{QaasHost: qaasHost, QaasPort: qaasPort, QaasCertFile: qaasCertFile}
		url, err := webhook.New(args[0])
		if err != nil {
			log.Errorf("fail creating new webhook: %+v\n", err)
			return
		}
		log.Infof("webhook created successfully with URL: %s\n", url.String())
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List availabile webhooks.",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		webhook := qaas.Webhook{QaasHost: qaasHost, QaasPort: qaasPort, QaasCertFile: qaasCertFile}
		ws, err := webhook.List()
		if err != nil {
			log.Errorf("fail retriving list of webhooks: %+v\n", err)
			return
		}
		for w := range ws {
			log.Infof("- %+v\n", w)
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete an existing webhook.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Warnf("Not implemented!!")
	},
}

var infoCmd = &cobra.Command{
	Use:   "info [name]",
	Short: "Retrieve information of an existing webhook.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Warnf("Not implemented!!")
	},
}

var triggerCmd = &cobra.Command{
	Use:   "trigger [name]",
	Short: "Trigger an existing webhook manually.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Warnf("Not implemented!!")
	},
}
