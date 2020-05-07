package ftp

import (
	"github.com/shawnburke/amcrest-viewer/common"

	ftps "github.com/goftp/server"
)

type ftpAuth struct {
	auth common.Auth
}

func createAuth(auth common.Auth) ftps.Auth {

	a := &ftpAuth{
		auth: auth,
	}
	return a
}

func (fa *ftpAuth) CheckPasswd(user string, pass string) (bool, error) {
	ok := fa.auth.IsAllowed(user, pass)
	return ok, nil
}
