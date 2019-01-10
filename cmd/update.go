package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tufclient "github.com/theupdateframework/notary/client"

	"github.com/foundriesio/tuftree/client"
)

var (
	baseVer        string
	personalityVer string
	updateCmd      = &cobra.Command{
		Use:   "update",
		Short: "Update the base image and/or personality of the device",
		Run:   doUpdate,
	}
)

func init() {
	RootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&baseVer, "base", "", "latest", "The version to update to. If set empty, no update will be performed")
	updateCmd.Flags().StringVarP(&personalityVer, "personality", "", "latest", "The version to update to. If set empty, no update will be performed")
}

func doUpdate(cmd *cobra.Command, args []string) {
	var base, personality *tufclient.TargetWithRole

	if device.BaseNotary == nil && len(baseVer) > 0 {
		logrus.Error("Device is not configured for base updates")
	} else if len(baseVer) > 0 {
		logrus.Info("Probing server for base updates")
		targets, err := device.BaseTargets()
		if err != nil {
			logrus.Error(err)
			return
		}
		for _, target := range targets {
			ver, _ := client.BaseVersionSplit(target.Name)
			if baseVer == "latest" || ver == baseVer {
				base = target
				break
			}
		}
		if base == nil {
			logrus.Fatal("Can't find base update")
		}
	}
	if device.PersonalityNotary == nil && len(personalityVer) > 0 {
		logrus.Error("Device is not configured for personality updates")
	} else if len(personalityVer) > 0 {
		logrus.Info("Probing server for personality updates")
		targets, err := device.PersonalityTargets()
		if err != nil {
			logrus.Error(err)
			return
		}
		for _, target := range targets {
			if personalityVer == "latest" || target.Name == personalityVer {
				personality = target
				break
			}
		}
		if personality == nil {
			logrus.Fatal("Can't find personality update")
		}
	}

	if base != nil {
		if err := device.UpdateBase(base); err != nil {
			logrus.Fatal(err)
		}
	}
	if personality != nil {
		if err := device.UpdatePersonality(personality); err != nil {
			logrus.Fatal(err)
		}
	}
}
