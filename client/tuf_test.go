package client

import (
	"testing"

	"github.com/docker/go/canonical/json"
)

func TestBadOStreeCustom(t *testing.T) {
	custom := json.RawMessage([]byte(`{
		"targetFormat": "OSTREE",
		"missing field": "ostree"
	}`))
	_, err := NotaryClient{}.OSTree(&custom)
	if err == nil {
		t.Error("OSTREE parsing should have failed")
	}

	custom = json.RawMessage([]byte(`{
		"targetFormat": "invalid",
		"ostree": "foo",
	}`))
	_, err = NotaryClient{}.OSTree(&custom)
	if err == nil {
		t.Error("OSTREE parsing should have failed")
	}
}

func TestGoodOSTreeCustom(t *testing.T) {
	custom := json.RawMessage([]byte(`{
		"targetFormat": "OSTREE",
		"ostree": "foo"
	}`))
	c, err := NotaryClient{}.OSTree(&custom)
	if err != nil {
		t.Fatalf("OSTREE parsing failed with: %s", err)
	}
	if c.Url != "foo" {
		t.Errorf("OSTREE URL %s != foo", c.Url)
	}
}

func TestBadDockerComposeCustom(t *testing.T) {
	custom := json.RawMessage([]byte(`{
		"targetFormat": "DOCKER_COMPOSE",
		"missing field": "tgz"
	}`))
	_, err := NotaryClient{}.DockerCompose(&custom)
	if err == nil {
		t.Error("DOCKER_COMPOSE parsing should have failed")
	}

	custom = json.RawMessage([]byte(`{
		"targetFormat": "OSTREE",
		"tgz": "foo",
	}`))
	_, err = NotaryClient{}.DockerCompose(&custom)
	if err == nil {
		t.Error("DOCKER_COMPOSE parsing should have failed")
	}
}

func TestGoodDockerComposeCustom(t *testing.T) {
	custom := json.RawMessage([]byte(`{
		"uri": "example.com",
		"targetFormat": "DOCKER_COMPOSE",
		"tgz": "foo",
		"compose-env": {
			"foo": "bar",
			"bam": "bang"
		}
	}`))
	c, err := NotaryClient{}.DockerCompose(&custom)
	if err != nil {
		t.Fatalf("DOCKER_COMPOSE parsing failed with: %s", err)
	}
	if c.TgzUrl != "foo" {
		t.Errorf("DOCKER_COMPOSE TgzUrl %s != foo", c.TgzUrl)
	}
	if c.Uri != "example.com" {
		t.Errorf("DOCKER_COMPOSE Uri %s != example.com", c.Uri)
	}
	if c.ComposeEnv["foo"] != "bar" {
		t.Errorf("DOCKER_COMPOSE env[foo] %s != bar", c.ComposeEnv["foo"])
	}
	if c.ComposeEnv["bam"] != "bang" {
		t.Errorf("DOCKER_COMPOSE env[bam] %s != bang", c.ComposeEnv["bam"])
	}
}
