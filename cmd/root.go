package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/foundriesio/tuftree/client"
)

var (
	cmdVerbose   bool
	cmdConfigDir string
	device       *client.Device
)

var RootCmd = &cobra.Command{
	Use:               "tuftree",
	Short:             "tuftree keeps base OS images and personalities up-to-date",
	PersistentPreRunE: initConfig,
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&cmdVerbose, "verbose", "v", false, "Print more information")
	RootCmd.PersistentFlags().StringVarP(&cmdConfigDir, "config-dir", "c", "/var/tuftree", "Configuration directory path to use")
}

func initConfig(cmd *cobra.Command, args []string) error {
	if cmdVerbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Debugf("Configuration location: %s", cmdConfigDir)

	if cmd == initializeCmd {
		return nil
	}
	var err error
	device, err = client.NewDevice(cmdConfigDir)
	if err != nil {
		logrus.Fatal(err)
	}
	return nil
}
