package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func NewOSTreeStatus() (*OSTreeStatus, error) {
	out, err := Run("ostree", "admin", "status")
	if err != nil {
		return nil, err
	}

	status := OSTreeStatus{}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			if fields[0] == "*" {
				idx := strings.Index(fields[2], ".")
				status.Active = fields[2][:idx]
			} else if len(fields) == 3 && fields[2] == "(pending)" {
				idx := strings.Index(fields[1], ".")
				sts := fields[1][:idx]
				status.Pending = &sts
			}
		}
	}
	return &status, nil
}

func OSTreeAddRemote(label string, url string, ignoreGPG bool) error {
	fd, err := os.Create("/etc/ostree/remotes.d/" + label + ".conf")
	if err != nil {
		return fmt.Errorf("Unable to create ostree remote config: %s", err)
	}
	defer fd.Close()
	fd.Chmod(0644)
	fd.WriteString("[remote \"" + label + "\"]\n")
	fd.WriteString("url=" + url + "\n")
	if ignoreGPG {
		fd.WriteString("gpg-verify=false\n")
	}
	return nil
}

func OSTreeUpdate(remote string, hash string) error {
	logrus.Infof("Pulling ostree objects for %s:%s", remote, hash)
	if err := RunStreamed("ostree", "pull", remote, hash); err != nil {
		return err
	}

	logrus.Infof("Deploying ostree image %s:%s", remote, hash)
	if err := RunStreamed("ostree", "admin", "deploy", hash); err != nil {
		return err
	}
	return nil
}
