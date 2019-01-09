package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/foundriesio/tuftree/client"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Display status of device",
		Run:   doStatus,
	}
)

func init() {
	RootCmd.AddCommand(statusCmd)
}

func doStatus(cmd *cobra.Command, args []string) {
	fmt.Printf("Hardware-id:\t%s\n", device.HardwareId)
	fmt.Printf("Active image:\t%s\n", device.OSTreeStatus.Active)
	if device.OSTreeStatus.Pending != nil {
		fmt.Printf("Pending image:\t%s\n", *device.OSTreeStatus.Pending)
	}

	if device.BaseNotary != nil {
		tgt, _, err := device.BaseTarget()
		if err != nil {
			fmt.Printf("Unable to find base version information: %s\n", err)
		} else {
			ver, _ := client.BaseVersionSplit(tgt.Name)
			fmt.Printf("Base Version:\t%s\n", ver)
		}
	}

	if device.PersonalityNotary != nil {
		tgt, _, err := device.PersonalityTarget()
		if err != nil {
			fmt.Printf("Unable to find personality version information: %s\n", err)
		} else {
			fmt.Printf("Personality Version:\t%s\n", tgt.Name)
		}
	}

}
