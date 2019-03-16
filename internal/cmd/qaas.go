package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	qaas "github.com/Donders-Institute/hpc-qaas/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var qaasHost string
var qaasPort int
var qaasCertFile string
var webhookName string
var webhookPayload string
var webhookPayloadType string

// variable may be set at the build time to fix the default location for the QaaS server certificate.
var defQaasCert string

func init() {

	qaasCmd.PersistentFlags().StringVarP(&qaasHost, "server", "s", "qaas.dccn.nl", "QaaS service hostname")
	qaasCmd.PersistentFlags().IntVarP(&qaasPort, "port", "p", 443, "QaaS service hostname")
	qaasCmd.PersistentFlags().StringVarP(&qaasCertFile, "cert", "c", defQaasCert, "QaaS service SSL certificate")

	createCmd.Flags().StringVarP(&webhookName, "name", "n", "", "name or a short description of the webhook")
	triggerCmd.Flags().StringVarP(&webhookPayload, "payload", "l", "", "file containing the webhook payload data")
	triggerCmd.Flags().StringVarP(&webhookPayloadType, "type", "t", "json", "webhook payload data type (json, xml or txt)")

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
		webhook := qaas.WebhookConfig{QaasHost: qaasHost, QaasPort: qaasPort, QaasCertFile: qaasCertFile}
		url, err := webhook.New(args[0], webhookName)
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
		webhook := qaas.WebhookConfig{QaasHost: qaasHost, QaasPort: qaasPort, QaasCertFile: qaasCertFile}
		ws, err := webhook.List()
		if err != nil {
			log.Errorf("fail retriving list of webhooks: %+v\n", err)
			return
		}
		for w := range ws {
			printWebhookConfigInfo(w)
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete an existing webhook.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		webhook := qaas.WebhookConfig{QaasHost: qaasHost, QaasPort: qaasPort, QaasCertFile: qaasCertFile}
		for _, id := range args {
			if err := webhook.Delete(id, true); err != nil {
				log.Errorln(err)
				continue
			}
			log.Infof("Webhook %s deleted.\n", id)
		}
	},
}

var infoCmd = &cobra.Command{
	Use:   "info [id]",
	Short: "Retrieve information of an existing webhook.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		webhook := qaas.WebhookConfig{QaasHost: qaasHost, QaasPort: qaasPort, QaasCertFile: qaasCertFile}
		for _, id := range args {
			if info, err := webhook.GetInfo(id); err != nil {
				log.Errorln(err)
				continue
			} else {
				printWebhookConfigInfo(info)
			}
		}
	},
}

var triggerCmd = &cobra.Command{
	Use:   "trigger [id]",
	Short: "Trigger webhook manually with a payload.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var dataPayload []byte

		// when payload is specified, check existence of it.
		if webhookPayload != "" {
			payloadAbs, err := filepath.Abs(webhookPayload)
			log.Fatalln(err)

			fi, err := os.Lstat(payloadAbs)
			if err != nil {
				log.Fatalln(err)
			}
			if !fi.Mode().IsRegular() {
				log.Fatalln("not a regular file: %s\n", payloadAbs)
			}
			dataPayload, err = ioutil.ReadFile(payloadAbs)
			if err != nil {
				log.Fatalln(err)
			}
		}

		// get webhook info
		webhook := qaas.WebhookConfig{QaasHost: qaasHost, QaasPort: qaasPort, QaasCertFile: qaasCertFile}
		info, err := webhook.GetInfo(args[0])
		if err != nil {
			log.Fatalln(err)
		}

		// map webhookPayloadType to request body content type
		// TODO: use a better approach (e.g. the mime package) with integration with bash completion.
		reqBodyType := "application/json"
		switch webhookPayloadType {
		case "xml":
			reqBodyType = "text/xml"
		case "txt":
			reqBodyType = "text/plain"
		default:
		}

		// make a POST call to the Webhook's URL with the content of payload as request body
		rspData, err := info.TriggerWebhook(dataPayload, reqBodyType, qaasCertFile)
		if err != nil {
			log.Fatalln(err)
		}
		log.Infof("webhook %s triggerd: %s\n", info.WebhookURL, string(rspData))
	},
}

// printWebhookConfigInfo writes one or multiple WebhookConfigInfo data objects to the stdout.
func printWebhookConfigInfo(infoList ...qaas.WebhookConfigInfo) {
	for _, info := range infoList {
		fmt.Printf("\n%-s", info.ID)
		fmt.Printf("\n\t%-16s:%-s", "Description", info.Description)
		fmt.Printf("\n\t%-16s:%-s", "Creation time", info.CreationTime)
		fmt.Printf("\n\t%-16s:%-s", "Script path", info.Script)
		fmt.Printf("\n\t%-16s:%-s", "Webhook URL", info.WebhookURL)
		fmt.Println()
	}
}
