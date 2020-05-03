package ftp

import (
	"io"

	filedriver "github.com/goftp/file-driver"
	"github.com/shawnburke/amcrest-viewer/common"
	"go.uber.org/zap"

	"fmt"

	ftps "github.com/shawnburke/amcrest-viewer/ftp-server"
)


type ftpFileSystem struct {
	server *ftps.Server
	dir string
	password string
	port int
	host string
	logger *zap.Logger
}

type FtpServer interface {
	Start() error
	Stop() error
}

func New(args *common.Params, logger *zap.Logger) FtpServer {
	fmt.Println("Created FTP server")
	return &ftpFileSystem{
		dir: args.DataDir,
		port: args.FtpPort,
		host: args.Host,
		password: args.FtpPassword,
		logger: logger,
	}
} 

func (fs *ftpFileSystem) Start() error {
	if fs.server != nil {
		return nil
	}
	factory := &fileDriverFactory{
		RootPath: fs.dir,
		Perm:     ftps.NewSimplePerm("user", "group"),
	}

	opts := &ftps.ServerOpts{
		Factory:  factory,
		Port:     fs.port,
		Hostname: fs.host,
		//Auth:     &ftps.SimpleAuth{Name: , Password: *pass},
	}

	fs.server = ftps.NewServer(opts)

	go func() {
	err := fs.server.ListenAndServe()
		if err != nil {
			fs.logger.Fatal("Error starting server:", err)
		}
	}()
	return nil
}

func (fs *ftpFileSystem) Stop() error {
	if fs.server != nil {
		server := fs.server
		fs.server = nil
		return server.Shutdown()
	}
	return nil
}

type fileDriverFactory struct {
	RootPath string
	ftps.Perm
}

func (factory *fileDriverFactory) NewDriver() (ftps.Driver, error) {
	fd := &filedriver.FileDriver{
		factory.RootPath,
		factory.Perm,
	}
	return &fileDriver{FileDriver:fd}, nil
}

type fileDriver struct {
	filedriver.FileDriver
}

func (fd *fileDriver) PutFile(destPath string, data io.Reader, appendData bool) (int64, error) {
	return fd.FileDriver.PutFile(destPath, data, appendData)
}


// AUTH