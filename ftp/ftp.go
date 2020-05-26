package ftp

import (
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
	auth     common.Auth
	bus      common.EventBus
}

type FtpServer interface {
	Start() error
	Stop() error
}

func New(
	args *common.Params,
	auth common.Auth,
	logger *zap.Logger,
	bus common.EventBus,
) FtpServer {
	return &ftpFileSystem{
		port:     args.FtpPort,
		host:     args.Host,
		password: args.FtpPassword,
		logger:   logger,
		auth:     auth,
		bus:      bus,
	}
}

func (fs *ftpFileSystem) Start() error {
	if fs.server != nil {
		return nil
	}
	factory := &fileDriverFactory{
		Perm:   ftps.NewSimplePerm("user", "group"),
		logger: fs.logger,
		bus:    fs.bus,
	}

	opts := &ftps.ServerOpts{
		Factory:  factory,
		Port:     fs.port,
		Hostname: fs.host,
		Auth:     createAuth(fs.auth, fs.bus, fs.logger),
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
