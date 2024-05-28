package ingest

import (
	"fmt"
	"io"

	"github.com/shawnburke/amcrest-viewer/cameras"
	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/ftp"
	"github.com/shawnburke/amcrest-viewer/storage"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/file"
	"github.com/shawnburke/amcrest-viewer/storage/models"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Invoke(NewIngestManager),
)

type IngestManagerParams struct {
	fx.In
	Logger      *zap.Logger
	Bus         common.EventBus
	FileManager file.Manager
	DataManager data.Repository
	Registry    cameras.Registry
}

type Ingester interface {
	Name() string
	IngestFile(cam *entities.Camera, f *ftp.File) (*models.MediaFile, error)
}

func NewIngestManager(p IngestManagerParams) error {
	im := &ingestManager{
		logger:    p.Logger,
		ingesters: map[string]Ingester{},
		fm:        p.FileManager,
		dm:        p.DataManager,
		registry:  p.Registry,
		bus:       p.Bus,
	}

	err := p.Bus.Subscribe(im)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %v", err)
	}

	return err
}

type ingestManager struct {
	logger    *zap.Logger
	ingesters map[string]Ingester
	fm        file.Manager
	dm        data.Repository
	registry  cameras.Registry
	bus       common.EventBus
}

func (im *ingestManager) getIngesterCamera(user string) (*entities.Camera, error) {

	cam, err := im.dm.GetCamera(user)
	if err != nil {
		im.logger.Error("Failed to load camera",
			zap.String("camera", user),
			zap.Error(err),
		)
		return nil, err
	}
	return cam, err
}

func (im *ingestManager) OnEvent(e common.Event) error {
	switch ev := e.(type) {
	case *ftp.FileCreateEvent:
		return im.ingestFtp(ev.File)
	case *ftp.FileRenameEvent:
		return im.ingestFtp(ev.File)
	case *storage.MediaFileAvailableEvent:
		return im.ingest(ev.File, ev.Reader)
	}
	return nil
}

// TODO: move this code into FTP package
func (im *ingestManager) ingestFtp(f *ftp.File) error {
	cam, err := im.getIngesterCamera(f.User)

	if err != nil {
		return err
	}

	if cam == nil {
		return fmt.Errorf("failed to find camera for user %q", f.User)
	}

	camType, err := im.registry.Get(cam.Type)

	if err != nil || camType == nil {
		im.logger.Error("Can't find ingester type",
			zap.String("type", cam.Type),
			zap.String("user", f.User),
			zap.String("file", f.FullName),
		)
		return nil
	}

	im.logger.Debug("Ingesting FTP file",
		zap.String("name", f.FullName),
		zap.String("User", f.User),
		zap.String("ingester", camType.Name()),
	)

	mf, err := camType.ParseFilePath(cam, f.FullName)

	if mf == nil {

		if err == common.ErrIngestIgnore {
			return nil
		}

		if err == common.ErrIngestDelete {
			return nil
		}

		im.logger.Error("Ingester failed to ingest file", zap.String("file", f.FullName), zap.Error(err))
		return nil
	}

	// Here we republish the event to the bus rather than calling ingest directly.
	// This is a bit more complicated but allows other systems to hook to this event,
	// and allows us to move the FTP stuff out w/o breaking this.
	//
	err = im.bus.Send(storage.NewMediaFileAvailableEvent(mf, f.Reader))
	if err != nil {
		im.logger.Error("Error sending new file to bus", zap.Error(err), zap.String("path", f.FullName))
		return err
	}
	return nil
}

func (im *ingestManager) ingest(mf *models.MediaFile, reader io.Reader) error {

	// make sure we always persist UTC
	mf.Timestamp = mf.Timestamp.UTC()

	// TODO: make manager interfaces speak models
	fileType := entities.FileTypeMp4

	switch mf.Type {
	case models.MP4:
		// fall through
	case models.JPG:
		fileType = entities.FileTypeJpg
		if mf.Duration == nil {
			freq := common.SnapshotFrequency
			mf.Duration = &freq
		}
	default:
		return fmt.Errorf("unknown file type: %v", mf.Type)
	}

	data, err := io.ReadAll(reader)
	relPath, err := im.fm.AddFile(mf.CameraID, data, mf.Timestamp, fileType)

	if err != nil {
		im.logger.Error("Failed to save file",
			zap.String("name", mf.Name), zap.String("camera", mf.CameraID), zap.Error(err))
		return fmt.Errorf("failed to safe file %v: %w", mf.Name, err)
	}

	fileData, err := im.dm.AddFile(relPath, fileType, mf.CameraID, len(data), mf.Timestamp, mf.Duration)

	if err != nil {
		im.logger.Error("Failed to save file data",
			zap.Error(err),
			zap.String("name", mf.Name),
			zap.String("camera", mf.CameraID),
			zap.String("disk-path", relPath),
			zap.Reflect("file", mf))
		return fmt.Errorf("failed to save file data: %w", err)
	}

	im.logger.Info("Ingested file",
		zap.Int("file-id", fileData.ID),
		zap.String("camera", mf.CameraID),
		zap.String("path", relPath))

	return nil
}
