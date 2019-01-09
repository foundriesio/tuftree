package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	listPersonalityCmd = &cobra.Command{
		Use:   "list-personality",
		Short: "List personality updates on server",
		Run:   doListPersonality,
	}
)

func init() {
	RootCmd.AddCommand(listPersonalityCmd)
}

func doListPersonality(cmd *cobra.Command, args []string) {
	if device.PersonalityNotary == nil {
		fmt.Println("Device is not configured for personality updates")
		return
	}
	fmt.Println("Updates:")
	targets, err := device.PersonalityTargets()
	if err != nil {
		logrus.Error(err)
		return
	}
	for _, target := range targets {
		hash := hex.EncodeToString(target.Hashes["sha256"])
		fmt.Printf("%s\t%s\n", target.Name, hash)
		c, err := device.BaseNotary.DockerCompose(target.Custom)
		if err != nil {
			logrus.Error(err)
		} else {
			fmt.Println("  TgzURL: ", c.TgzUrl)
			fmt.Println("  URL:    ", c.Uri)
		}
	}
}
