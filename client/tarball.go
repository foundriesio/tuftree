package client

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

func mkdir(header *tar.Header, dst string) error {
	dir := path.Join(dst, header.Name)
	err := os.Mkdir(dir, os.FileMode(header.Mode))
	if err != nil && os.IsNotExist(err) {
		return err
	}
	return nil
}

func mkfile(tr *tar.Reader, header *tar.Header, dst string) error {
	path := path.Join(dst, header.Name)
	fd, err := os.Create(path)
	defer fd.Close()
	if err != nil {
		return err
	}
	if err := fd.Chmod(os.FileMode(header.Mode)); err != nil {
		return err
	}
	_, err = io.CopyN(fd, tr, header.Size)
	return err
}

func mklink(header *tar.Header, dst string) error {
	to := path.Join(dst, header.Name)
	from := path.Join(path.Dir(to), header.Linkname)
	err := os.Link(from, to)
	if err != nil && os.IsNotExist(err) {
		return err
	}
	return nil
}

func mksymlink(header *tar.Header, dst string) error {
	to := path.Join(dst, header.Name)
	from := path.Join(path.Dir(to), header.Linkname)
	err := os.Symlink(from, to)
	if err != nil && os.IsNotExist(err) {
		return err
	}
	return nil
}

func extract(tr *tar.Reader, dst string, stripLeading bool) error {
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
		switch header.Typeflag {
		case tar.TypeReg:
			if err := mkfile(tr, header, dst); err != nil {
				return fmt.Errorf("Unable to extract file(%s): %s", header.Name, err)
			}
		case tar.TypeDir:
			if err := mkdir(header, dst); err != nil {
				return fmt.Errorf("Unable to make directory(%s): %s", header.Name, err)
			}
		case tar.TypeLink:
			if err := mklink(header, dst); err != nil {
				return fmt.Errorf("Unable to link %s -> %s: %s", header.Linkname, header.Name, err)
			}
		case tar.TypeSymlink:
			if err := mksymlink(header, dst); err != nil {
				return fmt.Errorf("Unable to symlink %s -> %s: %s", header.Linkname, header.Name, err)
			}
		case tar.TypeXGlobalHeader:
			continue
		default:
			logrus.Warnf("Unable to extract type: %d", header.Typeflag)
		}
	}
	return nil
}

func extractFile(fileName string, dst string, stripLeading bool) error {
	fd, err := os.Open(fileName)
	if err != nil {
		return err
	}
	gzf, err := gzip.NewReader(fd)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gzf)
	return extract(tr, dst, stripLeading)
}
