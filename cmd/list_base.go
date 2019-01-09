package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/foundriesio/tuftree/client"
)

var (
	listBaseCmd = &cobra.Command{
		Use:   "list-base",
		Short: "List base updates on server",
		Run:   doListBase,
	}
)

func init() {
	RootCmd.AddCommand(listBaseCmd)
}

func doListBase(cmd *cobra.Command, args []string) {
	if device.BaseNotary == nil {
		fmt.Println("Device is not configured for base updates")
		return
	}
	fmt.Println("Updates:")
	targets, err := device.BaseTargets()
	if err != nil {
		logrus.Error(err)
		return
	}
	for _, target := range targets {
		ver, _ := client.BaseVersionSplit(target.Name)
		hash := hex.EncodeToString(target.Hashes["sha256"])
		fmt.Printf("%s\t%s\n", ver, hash)
		c, err := device.BaseNotary.OSTree(target.Custom)
		if err != nil {
			logrus.Error(err)
		} else {
			fmt.Println("  OSTreeURL: ", c.Url)
			fmt.Println("  URL:       ", c.Uri)
		}
	}
}
