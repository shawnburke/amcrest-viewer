package storage

import (
	"io"
	"time"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/models"
)

const EventFileCreate = "media_file_available"

type MediaFileAvailableEvent struct {
	common.EventBase
	io.ReadCloser
	File *models.MediaFile
}

func (e *MediaFileAvailableEvent) Close() error {

	if e.ReadCloser != nil {
		return e.ReadCloser.Close()
	}
	return nil
}

func NewMediaFileAvailableEvent(mf *models.MediaFile, reader io.ReadCloser) *MediaFileAvailableEvent {
	return &MediaFileAvailableEvent{
		EventBase:  common.NewEventBase(EventFileCreate, time.Now()),
		ReadCloser: reader,
		File:       mf,
	}
}
