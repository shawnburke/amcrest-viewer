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
	factory := newDriverFactory(ftps.NewSimplePerm("user", "group"), fs.logger, fs.bus)

	opts := &ftps.ServerOpts{
		Factory:  factory,
		Port:     fs.port,
		Hostname: fs.host,
		Auth:     createAuth(fs.auth, fs.bus, fs.logger),
		Logger:   &zapLogger{logger: fs.logger, debug: true},
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

type zapLogger struct {
	logger *zap.Logger
	debug  bool
}

func (logger *zapLogger) log(sessionId string, format string, v ...interface{}) {
	sugared := logger.logger.Sugar()

	f := sugared.Infof

	if logger.debug {
		f = sugared.Debugf
	}

	v2 := []interface{}{sessionId}
	v2 = append(v2, v...)
	f("%s: "+format, v2...)
}

func (logger *zapLogger) Print(sessionId string, message interface{}) {

	logger.log(sessionId, "%s", message)
}

func (logger *zapLogger) Printf(sessionId string, format string, v ...interface{}) {
	logger.log(sessionId, format, v...)
}

func (logger *zapLogger) PrintCommand(sessionId string, command string, params string) {

	if command == "PASS" {
		logger.log(sessionId, "%> PASS ****")
	} else {
		logger.log(sessionId, " > %s %s", command, params)
	}
}

func (logger *zapLogger) PrintResponse(sessionId string, code int, message string) {
	logger.log(sessionId, " < %d %s", code, message)
}
