package ftp

import (
	"io"

	filedriver "github.com/goftp/file-driver"

	ftps "github.com/goftp/server"
)

type fileDriverFactory struct {
	RootPath string
	ftps.Perm
}

func (factory *fileDriverFactory) NewDriver() (ftps.Driver, error) {
	fd := filedriver.FileDriver{
		factory.RootPath,
		factory.Perm,
	}
	return &fileDriver{FileDriver: fd}, nil
}

type fileDriver struct {
	filedriver.FileDriver
}

func (fd *fileDriver) PutFile(destPath string, data io.Reader, appendData bool) (int64, error) {
	return fd.FileDriver.PutFile(destPath, data, appendData)
}
