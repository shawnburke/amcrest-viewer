package common

import (
	"io"
	"time"
)

// File should be renamed to FtpFile
type File struct {
	User       string
	Data       []byte
	Name       string
	FullName   string
	ReceivedAt time.Time
	Done       func()
}

func (f *File) Finish() {
	if f != nil && f.Done != nil {
		f.Done()
	}
}

const EventFileCreate = "file_create"
const EventFileRename = "file_rename"
const EventFileDelete = "file_delete"

type FileCreateEvent struct {
	EventBase
	File *File
}

func NewFileCreateEvent(f *File) *FileCreateEvent {
	return &FileCreateEvent{
		EventBase: NewEventBase(EventFileCreate, time.Now()),
		File:      f,
	}
}

type FileRenameEvent struct {
	EventBase
	File    *File
	OldName string
}

func NewFileRenameEvent(f *File, oldName string) *FileRenameEvent {
	return &FileRenameEvent{
		EventBase: NewEventBase(EventFileRename, time.Now()),
		File:      f,
		OldName:   oldName,
	}
}

type MediaFileType int

const (
	Unknown MediaFileType = 0
	MP4     MediaFileType = 1
	JPG     MediaFileType = 2
)

type MediaFile struct {
	ID         string
	Camera     Camera
	Type       MediaFileType
	Path       string
	Timestamp  time.Time
	Duration   *time.Duration
	ReceivedAt time.Time
}

func (f *File) Close() {
	if f.Done != nil {
		f.Done()
		f.Done = nil
	}
}

type Camera struct {
	ID   string
	Name string
	Type string
}

type FileRepository interface {
	Add(mf *MediaFile, f *File) error
	GetCameras() ([]Camera, error)
	GetMedia(cameraID string, start time.Time, end time.Time) ([]MediaFile, error)
	GetFile(id string) (io.ReadCloser, error)
}

