package ftp

import (
	"github.com/shawnburke/amcrest-viewer/common"
	"go.uber.org/zap"

	ftps "github.com/goftp/server"
)

type ftpAuth struct {
	auth common.Auth
	bus  common.EventBus
	logger *zap.Logger
}

func createAuth(auth common.Auth, bus common.EventBus, logger *zap.Logger) ftps.Auth {

	a := &ftpAuth{
		auth: auth,
		bus:  bus,
		logger:logger,
	}
	return a
}

func (fa *ftpAuth) CheckPasswd(user string, pass string) (bool, error) {
	ok := fa.auth.IsAllowed(user, pass)

	msg := "Login fail"
	if ok {
		msg = "Login success"
	}
	fa.logger.Info(msg, zap.String("user", user))
	

	if !ok {
		fa.bus.Send(common.NewAuthFailEvent(user))
		return false, nil
	
	fa.bus.Send(common.NewAuthEvent(user))
	return ok, nil
}
