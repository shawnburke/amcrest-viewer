package models

import (
	"time"
)

type FileDate struct {
	Date   time.Time
	Videos []*CameraVideo
	date   string
}

func (fd *FileDate) DateString() string {
	if fd.date == "" {
		fd.date = fd.Date.Format("2006-01-02")
	}
	return fd.date
}

type CameraFile struct {
	Time time.Time
	Path string
}

type CameraStill struct {
	CameraFile
}

type CameraVideo struct {
	CameraFile
	Duration time.Duration
	Images   []*CameraStill
}

func (cv CameraVideo) End() time.Time {
	return cv.Time.Add(cv.Duration)
}

type CameraItem interface {
	Timestamp() time.Time
	FilePath() string
}

func (cf *CameraFile) Timestamp() time.Time {
	return cf.Time
}

func (cf *CameraFile) FilePath() string {
	return cf.Path
}
