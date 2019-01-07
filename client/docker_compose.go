package client

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"
	"github.com/sirupsen/logrus"
)

func NewComposeUpdater(notaryUrl, cacheDir, hash string, dcc DockerComposeCustom) (*DockerComposeUpdater, error) {
	tgzFile := path.Join(cacheDir, hash) + ".tgz"
	if _, err := os.Stat(tgzFile); os.IsNotExist(err) {
		logrus.Infof("DOCKER_COMPOSE(%s) not cached locally, downloading now", hash)
		if err := downloadTo(tgzFile, dcc.TgzUrl, hash); err != nil {
			return nil, err
		}
	}

	reader, err := validateTgz(tgzFile, hash)
	if err != nil {
		return nil, err
	}
	composeFiles, err := composeFiles(dcc.TgzLeading, dcc.ComposeFiles, reader)
	if err != nil {
		return nil, err
	}

	if err := validateComposeImages(notaryUrl, composeFiles, dcc.ComposeEnv); err != nil {
		return nil, err
	}
	return &DockerComposeUpdater{cachedTgz: tgzFile, dcc: dcc}, nil
}

func (dcu *DockerComposeUpdater) Stop(projectDir string) error {
	return dcu.run(projectDir, "stop")
}

func (dcu *DockerComposeUpdater) Start(projectDir string) error {
	return dcu.run(projectDir, "up", "-d")
}

func (dcu *DockerComposeUpdater) run(projectDir string, args ...string) error {
	// Ensure our docker-compose directory has the files we expect
	logrus.Infof("Extracting docker-compose to %s", projectDir)
	if err := extractFile(dcu.cachedTgz, projectDir, dcu.dcc.TgzLeading); err != nil {
		return fmt.Errorf("Unable to extract docker-compose tarball: %s", err)
	}

	fileArgs := []string{}
	if len(dcu.dcc.ComposeFiles) == 0 {
		fileArgs = append(fileArgs, "-f", "docker-compose.yml")
	} else {
		for _, file := range dcu.dcc.ComposeFiles {
			fileArgs = append(fileArgs, "-f", file)
		}
	}
	args = append(fileArgs, args...)
	return RunFromStreamed(projectDir, "docker-compose", args...)
}

func downloadTo(dstFile, url, hash string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Unable to download %s : %s", url, err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Unable to read response from %s : %s", url, err)
	}
	sum := sha256.Sum256(buf)
	found := hex.EncodeToString(sum[:])
	if found != hash {
		return fmt.Errorf("Invalid sha256(%s) %s != %s", url, found, sum)
	}

	if err := ioutil.WriteFile(dstFile, buf, 0640); err != nil {
		return fmt.Errorf("Unable to create file %s : %s", dstFile, err)
	}
	return nil
}

func validateTgz(tgzFile, hash string) (*tar.Reader, error) {
	buf, err := ioutil.ReadFile(tgzFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read DOCKER_COMPOSE cache of %s: %s", tgzFile, err)
	}
	sum := sha256.Sum256(buf)
	found := hex.EncodeToString(sum[:])
	if found != hash {
		return nil, fmt.Errorf("DOCKER_COMPOSE cache changed on disk sha256(%s) %s != %s", tgzFile, found, hash)
	}

	gzf, err := gzip.NewReader(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("Unable to decompress DOCKER_COMPOSE cache %s: %s", tgzFile, err)
	}
	return tar.NewReader(gzf), nil
}

func composeFiles(stripLeading bool, composeFiles []string, tr *tar.Reader) ([]types.ConfigFile, error) {
	if len(composeFiles) == 0 {
		composeFiles = []string{"docker-compose.yml"}
	}
	var files []types.ConfigFile
	required := make(map[string]int)
	for _, name := range composeFiles {
		required[name] = 1
	}
	for true {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if stripLeading {
			idx := strings.Index(header.Name, "/")
			if idx > 0 {
				header.Name = header.Name[idx+1:]
			}
		}
		_, ok := required[header.Name]
		if ok {
			delete(required, header.Name)
			data := make([]byte, header.Size)
			_, err := tr.Read(data)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("Error reading %s from tgz data: %s", header.Name, err)
			}

			dict, err := loader.ParseYAML(data)
			if err != nil {
				return nil, fmt.Errorf("Invalid docker-compose(%s) in tgz data: %s", header.Name, err)
			}
			files = append(files, types.ConfigFile{Filename: header.Name, Config: dict})
		}
	}
	if len(required) > 0 {
		names := make([]string, len(required))
		for name, _ := range required {
			names = append(names, name)
		}
		return nil, fmt.Errorf("Missing required compose files in tgz data: %s", names)
	}
	return files, nil
}

func validateComposeImages(notaryUrl string, composeFiles []types.ConfigFile, env map[string]string) error {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	config := types.ConfigDetails{
		WorkingDir:  workingDir,
		ConfigFiles: composeFiles,
		Environment: env,
	}
	actual, err := loader.Load(config)
	if err != nil {
		return err
	}
	for _, svc := range actual.Services {
		if strings.HasPrefix(svc.Image, "hub.foundries.io") {
			logrus.Infof("Pulling/validating signed image: %s", svc.Image)
			if err := notaryPull(notaryUrl, svc.Image); err != nil {
				return err
			}
		} else {
			logrus.Debugf("Skipping notary validation of %s", svc.Image)
		}
	}
	return nil
}

func notaryPull(notaryUrl, image string) error {
	cmd := execCommand("docker", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"DOCKER_CONTENT_TRUST=1",
		"DOCKER_CONTENT_TRUST_SERVER="+notaryUrl,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Unable to run '%s': err=%s", cmd.Args, err)
	}
	return nil
}
