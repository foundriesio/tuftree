package client

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/docker/go/canonical/json"
	"github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary/client"
)

func DeviceInitialize(configDir string, config DeviceConfig) (*Device, error) {
	configFile := path.Join(configDir, "config.json")

	if len(config.HardwareId) == 0 {
		logrus.Info("Probing OSTree and Notary for Hardware ID")
		trustDir := path.Join(configDir, "notary")
		if err := os.MkdirAll(trustDir, 0700); err != nil {
			return nil, fmt.Errorf("Unable to create config-dir: %s", err)
		}
		tgt := probeTarget(config, trustDir)
		_, config.HardwareId = BaseVersionSplit(tgt.Name)

		data, err := json.Marshal(tgt)
		if err != nil {
			return nil, fmt.Errorf("Unable create base target: %s", err)
		}
		err = ioutil.WriteFile(path.Join(configDir, "base.json"), data, 0640)
		if err != nil {
			return nil, fmt.Errorf("Unable write base target: %s", err)
		}
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("Unable create configuration: %s", err)
	}

	err = ioutil.WriteFile(configFile, data, 0640)
	if err != nil {
		return nil, fmt.Errorf("Unable write configuration: %s", err)
	}

	return NewDevice(configDir)
}

func NewDevice(configDir string) (*Device, error) {
	configFile := path.Join(configDir, "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("'initialize' has not been run")
	}
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Fatalf("Error reading %s: %s", configFile, err)
	}
	config := DeviceConfig{}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		logrus.Fatalf("Error in %s: %s", configFile, err)
	}

	status, err := NewOSTreeStatus()
	if err != nil {
		return nil, err
	}

	d := Device{
		HardwareId:   config.HardwareId,
		configDir:    configDir,
		Config:       config,
		OSTreeStatus: status,
	}

	trustDir := path.Join(configDir, "notary")
	if len(config.BaseCollectionName) > 0 {
		d.BaseNotary = &NotaryClient{
			trustDir:   trustDir,
			serverURL:  config.BaseNotaryServerUrl,
			rootCAFile: config.BaseNotaryCAFile,
		}
	}
	if len(config.PersonalityCollectionName) > 0 {
		d.PersonalityNotary = &NotaryClient{
			trustDir:   trustDir,
			serverURL:  config.PersonalityNotaryServerUrl,
			rootCAFile: config.PersonalityNotaryCAFile,
		}
	}

	return &d, nil
}

func (d *Device) BaseTargets() ([]*client.TargetWithRole, error) {
	return d.BaseNotary.Targets(d.Config.BaseCollectionName)
}

func (d *Device) PersonalityTargets() ([]*client.TargetWithRole, error) {
	return d.PersonalityNotary.Targets(d.Config.PersonalityCollectionName)
}

func (d *Device) BaseTarget() (*client.TargetWithRole, *OSTreeCustom, error) {
	bytes, err := ioutil.ReadFile(path.Join(d.configDir, "base.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to find device's configured target: %s", err)
	}
	target := client.TargetWithRole{}
	if err := json.Unmarshal(bytes, &target); err != nil {
		return nil, nil, fmt.Errorf("Unable to parse device's configured target: %s", err)
	}
	if target.Custom == nil || target.Name == "" {
		return nil, nil, fmt.Errorf("Invalid base target data: %s", bytes)
	}
	ostree, err := NotaryClient{}.OSTree(target.Custom)
	if err != nil {
		return nil, nil, fmt.Errorf("Invalid OSTREE custom data: %s", err)
	}
	return &target, ostree, nil
}

func (d *Device) PersonalityTarget() (*client.TargetWithRole, *DockerComposeCustom, error) {
	bytes, err := ioutil.ReadFile(path.Join(d.configDir, "personality.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to find configured personality target: %s", err)
	}
	target := client.TargetWithRole{}
	if err := json.Unmarshal(bytes, &target); err != nil {
		return nil, nil, fmt.Errorf("Unable to parse configured personality target: %s", err)
	}
	if target.Custom == nil || target.Name == "" {
		return nil, nil, fmt.Errorf("Invalid base target data: %s", bytes)
	}
	dcc, err := NotaryClient{}.DockerCompose(target.Custom)
	if err != nil {
		return nil, nil, fmt.Errorf("Invalid DOCKER_COMPOSE custom data: %s", err)
	}
	return &target, dcc, nil
}

func (d *Device) UpdateBase(target *client.TargetWithRole) error {
	desired := hex.EncodeToString(target.Hashes["sha256"])
	if d.OSTreeStatus.Active == desired {
		logrus.Infof("Device already running ostree hash %s", desired)
		if err := saveTarget(path.Join(d.configDir, "base.json"), target); err != nil {
			return err
		}
		return nil
	}

	ver, hwid := BaseVersionSplit(target.Name)
	if hwid != d.HardwareId {
		logrus.Fatalf("Unexpected hardware id for this update: %s", hwid)
	}

	custom, err := d.BaseNotary.OSTree(target.Custom)
	if err != nil {
		return err
	}

	logrus.Infof("Updating device to version %s, ostree hash %s", ver, desired)
	if err := OSTreeAddRemote("tuftree", custom.Url, true); err != nil {
		return err
	}
	if err := OSTreeUpdate("tuftree", desired); err != nil {
		return err
	}
	return nil
}

func (d *Device) UpdatePersonality(target *client.TargetWithRole) error {
	desired := hex.EncodeToString(target.Hashes["sha256"])

	composeDir := path.Join(d.configDir, "docker-compose-current")
	if err := os.MkdirAll(composeDir, 0700); err != nil {
		return fmt.Errorf("Unable to create docker-compose directory: %s", err)
	}

	cacheDir := path.Join(d.configDir, "docker-compose-cache")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("Unable to create docker-compose cache: %s", err)
	}

	custom, err := d.PersonalityNotary.DockerCompose(target.Custom)
	if err != nil {
		return err
	}

	logrus.Infof("Updating personality to version %s, ostree hash %s", target.Name, desired)
	new, err := NewComposeUpdater(d.PersonalityNotary.serverURL, cacheDir, desired, *custom)
	if err != nil {
		return err
	}

	oldTgt, custom, err := d.PersonalityTarget()
	if err != nil {
		logrus.Warnf("Error loading current personality, assuming initial run: %s", err)
	} else {
		hash := hex.EncodeToString(oldTgt.Hashes["sha256"])
		old, err := NewComposeUpdater(d.PersonalityNotary.serverURL, cacheDir, hash, *custom)
		if err != nil {
			logrus.Warnf("Unable to load old personality, skipping docker-compose-stop: %s", err)
		} else {
			if err := old.Stop(composeDir); err != nil {
				logrus.Warnf("Unable to stop old personality, continuing with fingers crossed: %s", err)
			}
		}
	}

	if err := new.Start(composeDir); err != nil {
		return fmt.Errorf("Unable to start new personality: %s", err)
	}
	return nil
}

// Takes a target name from a Base image collection like v38-hikey
// and returns a tuple(version, hardwareId)
func BaseVersionSplit(targetName string) (string, string) {
	idx := strings.Index(targetName, "-")
	if idx < 1 {
		logrus.Fatalf("Invalid target name: %s. Must be formatted as <version>-<hardwareId>", targetName)
	}
	return targetName[:idx], targetName[idx+1:]
}

func probeTarget(config DeviceConfig, trustDir string) *client.TargetWithRole {
	notary := NotaryClient{
		trustDir:   trustDir,
		serverURL:  config.BaseNotaryServerUrl,
		rootCAFile: config.BaseNotaryCAFile,
	}
	targets, err := notary.Targets(config.BaseCollectionName)
	if err != nil {
		logrus.Fatalf("Unable to probe hardware ID, you'll need to set this manually: error=%s", err)
	}

	status, err := NewOSTreeStatus()
	if err != nil {
		logrus.Fatalf("Unable to probe hardware ID, you'll need to set this manually: error=%s", err)
	}

	for _, target := range targets {
		hash := hex.EncodeToString(target.Hashes["sha256"])
		if hash == status.Active {
			return target
		}
	}
	logrus.Fatalf("Unable to find device's hash(%s) in known updates", status.Active)
	return nil
}
