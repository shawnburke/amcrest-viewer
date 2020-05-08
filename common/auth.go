package common

import (
	"strings"
	"time"

	fxcfg "go.uber.org/config"
)

type Auth interface {
	IsAllowed(user, pwd string) bool
}

func NewConfigAuth(cfg fxcfg.Provider) (Auth, error) {
	ca := &configAuth{
		users: map[string]string{},
	}

	err := cfg.Get("ftp.users").Populate(ca.users)

	if err != nil {
		return nil, err
	}

	for k, v := range ca.users {
		noCaseKey := strings.ToLower(k)
		delete(ca.users, k)
		ca.users[noCaseKey] = v
	}

	return ca, nil

}

type configAuth struct {
	users map[string]string
}

func (ca *configAuth) IsAllowed(user, pwd string) bool {
	user = strings.ToLower(user)
	pass, ok := ca.users[user]
	return ok && pass == pwd
}

const EventUserLogin = "User_auth"
const EventUserLoginFail = "user_auth_fail"

type UserAuthFailEvent struct {
	EventBase
	User string
}

func NewAuthFailEvent(user string) *UserAuthFailEvent {
	return &UserAuthFailEvent{
		EventBase: NewEventBase(EventUserLoginFail, time.Now()),
		User:      user,
	}
}

func NewAuthEvent(user string) *UserAuthFailEvent {
	return &UserAuthFailEvent{
		EventBase: NewEventBase(EventUserLogin, time.Now()),
		User:      user,
	}
}
