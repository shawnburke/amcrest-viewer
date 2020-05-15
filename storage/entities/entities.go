package entities

import "time"

type Camera struct {
	ID       int        `db:"ID"`
	Name     string     `db:"Name"`
	Type     string     `db:"Type"`
	Host     *string    `db:"Host"`
	LastSeen *time.Time `db:"LastSeen"`
	Enabled  bool       `db:"Enabled"`
}

type File struct {
	ID              int       `db:"LastSeen"`
	CameraID        int       `db:"CameraID"`
	Path            string    `db:"Path"`
	Type            int       `db:"Type"`
	Timestamp       time.Time `db:"Timestamp"`
	DurationSeconds int       `db:"DurationSeconds"`
}

func (f File) Duration() time.Duration {
	return time.Second * time.Duration(f.DurationSeconds)
}
