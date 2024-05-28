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
	File   *models.MediaFile
	Reader io.Reader
}

func NewMediaFileAvailableEvent(mf *models.MediaFile, reader io.Reader) *MediaFileAvailableEvent {
	return &MediaFileAvailableEvent{
		EventBase: common.NewEventBase(EventFileCreate, time.Now()),
		File:      mf,
		Reader:    reader,
	}
}
