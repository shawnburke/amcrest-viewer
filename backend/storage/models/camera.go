package models

import (
	"time"

	"github.com/shawnburke/amcrest-viewer/storage/entities"
)

type Camera struct {
	ID       string
	Name     string
	Type     string
	Host     string
	LastSeen time.Time
	Enabled  bool
}

func FromCamera(cam *entities.Camera) *Camera {
	c := &Camera{
		ID:   cam.CameraID(),
		Name: cam.Name,
		Type: cam.Type,
	}

	if cam.Host != nil {
		c.Host = *cam.Host
	}

	if cam.LastSeen != nil {
		c.LastSeen = *cam.LastSeen
	}

	return c
}
