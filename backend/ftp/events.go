package ftp

import (
	"time"

	"github.com/shawnburke/amcrest-viewer/common"
)

const EventFileCreate = "file_create"
const EventFileRename = "file_rename"
const EventFileDelete = "file_delete"

type FileCreateEvent struct {
	common.EventBase
	File    *File
	NoClose bool
}

func (e *FileCreateEvent) Close() error {
	if e.NoClose {
		return nil
	}
	return e.File.Close()
}

func NewFileCreateEvent(f *File) *FileCreateEvent {
	return &FileCreateEvent{
		EventBase: common.NewEventBase(EventFileCreate, time.Now()),
		File:      f,
	}
}

type FileRenameEvent struct {
	common.EventBase
	File    *File
	OldName string
	NoClose bool
}

func (e *FileRenameEvent) Close() error {
	if e.NoClose {
		return nil
	}
	return e.File.Close()
}

func NewFileRenameEvent(f *File, oldName string) *FileRenameEvent {
	return &FileRenameEvent{
		EventBase: common.NewEventBase(EventFileRename, time.Now()),
		File:      f,
		OldName:   oldName,
	}
}
