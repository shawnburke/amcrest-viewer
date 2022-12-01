package common

import (
	"time"

	"github.com/shawnburke/amcrest-viewer/common"
)

type CameraCredsChangeEvent struct {
	common.EventBase
	CameraID string
}

func NewCameraCredsChangeEvent(cameraID string) *CameraCredsChangeEvent {
	return &CameraCredsChangeEvent{
		EventBase: common.NewEventBase("camera_creds_change", time.Now()),
		CameraID:  cameraID,
	}
}
