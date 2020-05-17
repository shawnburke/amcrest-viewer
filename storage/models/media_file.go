package models

import "time"

type MediaFileType int

const (
	Unknown MediaFileType = 0
	MP4     MediaFileType = 1
	JPG     MediaFileType = 2
)

type MediaFile struct {
	ID        string
	CameraID  string
	Type      MediaFileType
	Path      string
	Timestamp time.Time
	Duration  *time.Duration
}
