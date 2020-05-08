package ftp

import (
	"github.com/shawnburke/amcrest-viewer/common"

	ftps "github.com/goftp/server"
)

type ftpAuth struct {
	auth common.Auth
	bus  common.EventBus
}

func createAuth(auth common.Auth, bus common.EventBus) ftps.Auth {

	a := &ftpAuth{
		auth: auth,
		bus:  bus,
	}
	return a
}

func (fa *ftpAuth) CheckPasswd(user string, pass string) (bool, error) {
	ok := fa.auth.IsAllowed(user, pass)

	if !ok {
		fa.bus.Send(common.NewAuthFailEvent(user))
	} else {
		fa.bus.Send(common.NewAuthEvent(user))
	}
	return ok, nil
}
