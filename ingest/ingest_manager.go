package ingest

import (
	"errors"
	"fmt"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/ftp"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/file"
	"github.com/shawnburke/amcrest-viewer/storage/models"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Provide(Amcrest),
	fx.Invoke(NewIngestManager),
)

type IngestManagerParams struct {
	fx.In
	Logger      *zap.Logger
	Bus         common.EventBus
	Ingesters   []Ingester `group:"ingester"`
	FileManager file.Manager
	DataManager data.Repository
}

var ErrIngestDelete = errors.New("IngestDelete")
var ErrIngestIgnore = errors.New("IngestIgnore")

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
	}

	err := p.Bus.Subscribe(im)
	if err != nil {
		return fmt.Errorf("Failed to subscribe: %v", err)
	}

	for _, i := range p.Ingesters {
		im.ingesters[i.Name()] = i
	}
	return err
}

type ingestManager struct {
	logger    *zap.Logger
	ingesters map[string]Ingester
	fm        file.Manager
	dm        data.Repository
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
		return im.ingest(ev.File)
	case *ftp.FileRenameEvent:
		return im.ingest(ev.File)
	}
	return nil
}

func (im *ingestManager) ingest(f *ftp.File) error {
	cam, err := im.getIngesterCamera(f.User)

	if err != nil {
		return err
	}

	if cam == nil {
		return fmt.Errorf("Failed to find camera for user %q", f.User)
	}
	ingesterType := cam.Type

	ingester, ok := im.ingesters[ingesterType]

	if !ok {
		im.logger.Error("Can't find ingester type",
			zap.String("type", ingesterType),
			zap.String("user", f.User),
			zap.String("file", f.FullName),
		)
		return nil
	}

	im.logger.Info("Ingesting file",
		zap.String("name", f.FullName),
		zap.String("User", f.User),
		zap.String("ingester", ingester.Name()),
	)

	mf, err := ingester.IngestFile(cam, f)

	if mf == nil {

		if err == ErrIngestIgnore {
			return nil
		}

		if err == ErrIngestDelete {
			f.Close()
			return nil
		}

		im.logger.Error("Ingester failed to ingest file", zap.String("file", f.FullName), zap.Error(err))
		return nil
	}

	// make sure we always persist UTC
	mf.Timestamp = mf.Timestamp.UTC()

	im.logger.Info("Would ingest", zap.Reflect("media-file", mf))

	// TODO: make manager interfaces speak models
	fileType := entities.FileTypeMp4

	switch mf.Type {
	case models.MP4:
		// fall through
	case models.JPG:
		fileType = entities.FileTypeJpg
	default:
		return fmt.Errorf("Unknown file type: %v", mf.Type)
	}

	relPath, err := im.fm.AddFile(mf.CameraID, f.Data, mf.Timestamp, fileType)

	if err != nil {
		im.logger.Error("Failed to save file",
			zap.String("name", f.Name), zap.String("camera", mf.CameraID), zap.Error(err))
		return fmt.Errorf("Failed to safe file %v: %w", f.FullName, err)
	}

	fileData, err := im.dm.AddFile(relPath, fileType, mf.CameraID, len(f.Data), mf.Timestamp, mf.Duration)

	if err != nil {
		f2 := *f
		f2.Data = nil
		im.logger.Error("Failed to save file data",
			zap.Error(err),
			zap.String("name", f.Name),
			zap.String("camera", mf.CameraID),
			zap.String("disk-path", relPath),
			zap.Reflect("file", f2))
		return fmt.Errorf("Failed to save file data: %w", err)
	}

	im.logger.Info("Ingested file",
		zap.Int("file-id", fileData.ID),
		zap.String("camera", mf.CameraID),
		zap.String("path", relPath))

	f.Close()
	return nil
}
