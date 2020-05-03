package ftp

import (
	"strings"

	ftps "github.com/goftp/server"
)

type User struct {
	Name     string
	Password string
}

type ftpAuth struct {
	users []User
}

func createAuth() ftps.Auth {

	auth := &ftpAuth{
		users: []User{
			{
				"user1", "password1",
			},
			{
				"user2", "password2",
			},
		},
	}
	return auth
}

func (fa *ftpAuth) CheckPasswd(user string, pass string) (bool, error) {
	for _, u := range fa.users {
		if strings.EqualFold(u.Name, user) && u.Password == pass {
			return true, nil
		}
	}
	return false, nil
}
