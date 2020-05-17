package models

import (
	"time"
)

type Camera struct {
	ID       string
	Type     string
	Host     string
	LastSeen time.Time
	Enabled  bool
}
