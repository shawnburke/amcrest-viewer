package entities

import "time"

type Camera struct {
	ID       int        `db:"ID"`
	Name     string     `db:"Name"`
	Type     string     `db:"Type"`
	Host     *string    `db:"Host"`
	LastSeen *time.Time `db:"LastSeen"`
	Enabled  *bool       `db:"Enabled"`
}


const (
	FileTypeJpg int = 0
	FileTypeMp4 int = 1
)

type File struct {
	ID              int       `db:"ID"`
	CameraID        int       `db:"CameraID"`
	Path            string    `db:"Path"`
	Type            int       `db:"Type"`
	Timestamp       time.Time `db:"Timestamp"`
	Received		*time.Time `db:"Received"`
	DurationSeconds *int       `db:"DurationSeconds"`
}

func (f File) Duration() time.Duration {
	if f.DurationSeconds == nil {
		return time.Duration(0)
	}
	return time.Second * time.Duration(*f.DurationSeconds)
}
