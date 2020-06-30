package common


import (
	
	"io"
	
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/models"
	
)

type Type interface {

	Name() string
	Capabilities() Capabilities
	ParseFilePath(cam *entities.Camera, p string) (*models.MediaFile, error)
	Snapshot(cam *entities.Camera) (io.ReadCloser, error)
}