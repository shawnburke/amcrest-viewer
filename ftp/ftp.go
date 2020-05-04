package ftp

import (
	"fmt"
	"time"

	"github.com/shawnburke/amcrest-viewer/common"
	"go.uber.org/zap"

	ftps "github.com/goftp/server"
)

type ftpFileSystem struct {
	server   *ftps.Server
	dir      string
	password string
	port     int
	host     string
	logger   *zap.Logger
}

type FtpServer interface {
	Start() error
	Stop() error
}

func New(args *common.Params, logger *zap.Logger) FtpServer {
	fmt.Println("Created FTP server")
	return &ftpFileSystem{
		dir:      args.DataDir,
		port:     args.FtpPort,
		host:     args.Host,
		password: args.FtpPassword,
		logger:   logger,
	}
}

func (fs *ftpFileSystem) Start() error {
	if fs.server != nil {
		return nil
	}
	factory := &fileDriverFactory{
		RootPath: fs.dir,
		Perm:     ftps.NewSimplePerm("user", "group"),
		logger: fs.logger,
	}

	opts := &ftps.ServerOpts{
		Factory:  factory,
		Port:     fs.port,
		Hostname: fs.host,
		Auth:     createAuth(),
	}

	fs.server = ftps.NewServer(opts)

	// TODO: Clean up this mess with a better way to detect
	// clean startup
	var err error
	var ok bool
	go func() {
		err = fs.server.ListenAndServe()
		if err != nil && !ok {
			fs.logger.Fatal("Error starting server:", zap.Error(err))
		}
	}()
	time.Sleep(100 * time.Millisecond)
	if err == nil {
		ok = true
	}
	return err
}

func (fs *ftpFileSystem) Stop() error {
	if fs.server != nil {
		server := fs.server
		fs.server = nil
		return server.Shutdown()
	}
	return nil
}
