package client

import (
	"fmt"
	"net"
	"net/http"
	"sort"
	"time"

	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/go/canonical/json"
	"github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf/data"
)

func sortTargets(targets []*client.TargetWithRole) {
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Name > targets[j].Name
	})
}

func (c NotaryClient) Targets(image string) ([]*client.TargetWithRole, error) {
	gun := data.GUN(image)
	transport, err := c.getTransport(gun)
	if err != nil {
		logrus.Panicf("Unable to create notary transport: %s", err)
	}
	repo, err := client.NewFileCachedRepository(
		c.trustDir,
		gun,
		c.serverURL,
		transport,
		nil,
		trustpinning.TrustPinConfig{},
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to create notary cache for %s: %s", image, err.Error())
	}

	targets, err := repo.ListTargets()
	if err != nil {
		return nil, fmt.Errorf("Unable to list targets for %s: %s", image, err.Error())
	}

	sortTargets(targets)
	return targets, nil
}

func (c NotaryClient) getTransport(gun data.GUN) (http.RoundTripper, error) {
	tlsConfig, err := tlsconfig.Client(tlsconfig.Options{
		CAFile:             c.rootCAFile,
		ExclusiveRootPools: true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to configure TLS: %s", err.Error())
	}

	base := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
		DisableKeepAlives:   true,
	}

	modifiers := []transport.RequestModifier{
		transport.NewHeaderRequestModifier(http.Header{
			"User-Agent": []string{"tuftree"},
		}),
	}
	authTransport := transport.NewTransport(base, modifiers...)
	pingClient := &http.Client{
		Transport: authTransport,
		Timeout:   5 * time.Second,
	}
	req, err := http.NewRequest("GET", c.serverURL+"/v2/", nil)
	if err != nil {
		panic(err)
	}

	challengeManager := challenge.NewSimpleManager()
	resp, err := pingClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if err := challengeManager.AddResponse(resp); err != nil {
		panic(err)
	}
	tokenHandler := auth.NewTokenHandler(base, nil, gun.String(), "pull")
	modifiers = append(modifiers, auth.NewAuthorizer(challengeManager, tokenHandler, auth.NewBasicHandler(nil)))
	return transport.NewTransport(base, modifiers...), nil
}

func (c NotaryClient) OSTree(custom *json.RawMessage) (*OSTreeCustom, error) {
	otc := OSTreeCustom{}
	if custom != nil {
		err := json.Unmarshal(*custom, &otc)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse OSTREE custom data: %s", err)
		}
		if otc.TargetFormat != "OSTREE" {
			return nil, fmt.Errorf("Invalid targetFormat %s != OSTREE", otc.TargetFormat)
		}
		if len(otc.Url) == 0 {
			return nil, fmt.Errorf("Unable to parse OSTREE data, missing required filed 'ostree'")
		}
	}
	return &otc, nil
}

func (c NotaryClient) DockerCompose(custom *json.RawMessage) (*DockerComposeCustom, error) {
	dcc := DockerComposeCustom{}
	if custom != nil {
		err := json.Unmarshal(*custom, &dcc)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse DOCKER_COMPOSE custom data: %s", err)
		}
		if dcc.TargetFormat != "DOCKER_COMPOSE" {
			return nil, fmt.Errorf("Invalid targetFormat %s != DOCKER_COMPOSE", dcc.TargetFormat)
		}
		if len(dcc.TgzUrl) == 0 {
			return nil, fmt.Errorf("Unable to parse DOCKER_COMPOSE data, missing required filed 'tgz'")
		}
	}
	return &dcc, nil
}
