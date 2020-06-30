package storage

import (
	"time"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/models"
)

const EventFileCreate = "media_file_available"

type MediaFileAvailableEvent struct {
	common.EventBase
	File *models.MediaFile
	Data []byte
}

func NewMediaFileAvailableEvent(mf *models.MediaFile, data []byte) *MediaFileAvailableEvent {
	return &MediaFileAvailableEvent{
		EventBase: common.NewEventBase(EventFileCreate, time.Now()),
		File:      mf,
		Data:      data,
	}
}
