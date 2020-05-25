package entities

import (
	"fmt"
	"time"
)

type Camera struct {
	ID       int        `db:"ID" json:"id"`
	Name     string     `db:"Name" json:"name"`
	Type     string     `db:"Type" json:"type"`
	Host     *string    `db:"Host"  json:"host"`
	LastSeen *time.Time `db:"LastSeen"  json:"last_seen"`
	Enabled  *bool      `db:"Enabled" json:"enabled"`
}

func (c Camera) CameraID() string {
	return fmt.Sprintf("%s-%d", c.Type, c.ID)
}

const (
	FileTypeJpg int = 0
	FileTypeMp4 int = 1
)

type File struct {
	ID              int        `db:"ID"`
	CameraID        int        `db:"CameraID"`
	Path            string     `db:"Path"`
	Type            int        `db:"Type"`
	Timestamp       time.Time  `db:"Timestamp"`
	Received        *time.Time `db:"Received"`
	DurationSeconds *int       `db:"DurationSeconds"`
	Length          int        `db:"Length"`
}

func (f File) Duration() time.Duration {
	if f.DurationSeconds == nil {
		return time.Duration(0)
	}
	return time.Second * time.Duration(*f.DurationSeconds)
}
