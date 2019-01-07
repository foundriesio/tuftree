package client

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestBaseVersionSplit(t *testing.T) {
	ver, hwid := BaseVersionSplit("v123-intel")
	if ver != "v123" {
		t.Errorf("Invalid version %s != v123", ver)
	}
	if hwid != "intel" {
		t.Errorf("Invalid hwid %s != intel", hwid)
	}
}

func TestBaseTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "device-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	buf := []byte(`{
		"Name": "v123",
		"Hashes": {"sha256": "DEADBEEF"},
		"Length": 0,
		"Custom": {
			"targetFormat": "OSTREE",
			"ostree": "http://example.com"
		}

	}`)
	if err := ioutil.WriteFile(path.Join(dir, "base.json"), buf, 0700); err != nil {
		t.Fatal(err)
	}

	d := Device{configDir: dir}
	tgt, ostree, err := d.BaseTarget()
	if err != nil {
		t.Fatalf("Unable to parse base.json: %s", err)
	}
	if tgt.Name != "v123" {
		t.Errorf("Invalid target name: %s != v123", tgt.Name)
	}
	if ostree.Url != "http://example.com" {
		t.Errorf("Invalid ostree url: %s != http://example.com", ostree.Url)
	}
}

func TestPersonalityTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "device-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	buf := []byte(`{
		"Name": "v123",
		"Hashes": {"sha256": "DEADBEEF"},
		"Length": 0,
		"Custom": {
			"targetFormat": "DOCKER_COMPOSE",
			"tgz": "http://example.com"
		}

	}`)
	if err := ioutil.WriteFile(path.Join(dir, "personality.json"), buf, 0700); err != nil {
		t.Fatal(err)
	}

	d := Device{configDir: dir}
	tgt, dcc, err := d.PersonalityTarget()
	if err != nil {
		t.Fatalf("Unable to parse base.json: %s", err)
	}
	if tgt.Name != "v123" {
		t.Errorf("Invalid target name: %s != v123", tgt.Name)
	}
	if dcc.TgzUrl != "http://example.com" {
		t.Errorf("Invalid tgz url: %s != http://example.com", dcc.TgzUrl)
	}
}
