package client

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"
)

func createTgz(t *testing.T, contents map[string]string) (string, string) {
	tmpfile, err := ioutil.TempFile("", "docker-compose-tests")
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(tmpfile)
	tw := tar.NewWriter(gz)
	for name, content := range contents {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()
	gz.Close()
	tmpfile.Close()

	buf, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(buf)
	return tmpfile.Name(), hex.EncodeToString(sum[:])
}

func TestValidateTgz(t *testing.T) {
	contents := map[string]string{"foo": "bar"}
	tgz, hash := createTgz(t, contents)
	defer os.Remove(tgz)

	_, err := validateTgz(tgz, hash+"x")
	if err == nil {
		t.Error("validateTgz should fail with hash-mismatch")
	} else {
		t.Logf("Error message: %s", err)
	}
	_, err = validateTgz(tgz, hash)
	if err != nil {
		t.Errorf("validateTgz failed: %s", err)
	}
}

func TestComposeFiles(t *testing.T) {
	contents := map[string]string{"foo/blah": "{}"}
	tgz, hash := createTgz(t, contents)
	defer os.Remove(tgz)

	tr, err := validateTgz(tgz, hash)
	if err != nil {
		t.Fatalf("validateTgz failed: %s", err)
	}

	files := []string{}
	_, err = composeFiles(false, files, tr)
	if err == nil {
		t.Error("composeFiles should have failed with missing required file")
	} else {
		t.Logf("Error message: %s", err)
	}

	tr, err = validateTgz(tgz, hash)
	if err != nil {
		t.Fatalf("validateTgz failed: %s", err)
	}
	files = []string{"blah"}
	_, err = composeFiles(true, files, tr)
	if err != nil {
		t.Errorf("composeFiles failed: %s", err)
	}
}
