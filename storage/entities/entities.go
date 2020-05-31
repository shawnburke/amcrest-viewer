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
	ID              int        `db:"ID" json:"id"`
	CameraID        int        `db:"CameraID" json:"camera_id"`
	Path            string     `db:"Path" json:"path"`
	Type            int        `db:"Type" json:"type"`
	Timestamp       time.Time  `db:"Timestamp" json:"timestamp"`
	Received        *time.Time `db:"Received" json:"received_at"`
	DurationSeconds *int       `db:"DurationSeconds" json:"duration_seconds"`
	Length          int        `db:"Length" json:"length"`
}

func (f File) Duration() time.Duration {
	if f.DurationSeconds == nil {
		return time.Duration(0)
	}
	return time.Second * time.Duration(*f.DurationSeconds)
}
