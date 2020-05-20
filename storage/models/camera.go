package models

import (
	"time"
)

type Camera struct {
	ID       string
	Name     string
	Type     string
	Host     string
	LastSeen time.Time
	Enabled  bool
}
