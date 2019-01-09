package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/foundriesio/tuftree/client"
)

var (
	deviceConfig  = client.DeviceConfig{}
	initializeCmd = &cobra.Command{
		Use:   "initialize",
		Short: "Set up initial configuration",
		Run:   doInitialize,
	}
)

func init() {
	RootCmd.AddCommand(initializeCmd)

	initializeCmd.Flags().StringVarP(&deviceConfig.BaseNotaryServerUrl, "base-notary", "", "https://notary.foundries.io", "The notary server to use")
	initializeCmd.Flags().StringVarP(&deviceConfig.BaseCollectionName, "base-notary-collection", "", "hub.foundries.io/lmp", "The notary collection providing OSTree images")
	initializeCmd.Flags().StringVarP(&deviceConfig.BaseNotaryCAFile, "base-notary-ca", "", "", "Use an additional CA for talking to the server")

	initializeCmd.Flags().StringVarP(&deviceConfig.PersonalityNotaryServerUrl, "personality-notary", "", "https://notary.foundries.io", "The notary server to use")
	initializeCmd.Flags().StringVarP(&deviceConfig.PersonalityCollectionName, "personality-collection", "", "", "The notary collection providing DOCKER_COMPOSE details. If empty, no personality will be configured")
	initializeCmd.Flags().StringVarP(&deviceConfig.PersonalityNotaryCAFile, "personality-notary-ca", "", "", "Use an additional CA for talking to the server")

}

func doInitialize(cmd *cobra.Command, args []string) {
	fmt.Println("Initializing device state ...")
	d, err := client.DeviceInitialize(cmdConfigDir, deviceConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Hardware-id:\t%s\n", d.HardwareId)
	fmt.Printf("Active image:\t%s\n", d.OSTreeStatus.Active)
	if d.OSTreeStatus.Pending != nil {
		fmt.Printf("Pending image: %s\n", *d.OSTreeStatus.Pending)
	}
}
