package ingest

import (
	"fmt"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/ftp"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/file"
	"github.com/shawnburke/amcrest-viewer/storage/models"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Invoke(NewIngestManager),
	fx.Provide(Amcrest),
)

type IngestManagerParams struct {
	fx.In
	Logger      *zap.Logger
	Bus         common.EventBus
	Ingesters   []Ingester `group:"ingester"`
	FileManager file.Manager
	DataManager data.Repository
}

type Ingester interface {
	Name() string
	IngestFile(f *ftp.File) *models.MediaFile
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

func (im *ingestManager) getIngesterType(user string) string {
	return amcrestIngesterType
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
	ingesterType := im.getIngesterType(f.User)

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
		zap.String("name", f.Name),
		zap.String("User", f.User),
		zap.String("ingester", ingester.Name()),
	)

	mf := ingester.IngestFile(f)

	if mf == nil {
		im.logger.Error("Ingester failed to ingest file", zap.String("file", f.FullName))
		return nil
	}

	im.logger.Info("Would ingest", zap.Reflect("media-file", mf))

	// fileType := entities.FileTypeMp4

	// switch mf.Type {
	// case common.MP4:
	// case common.JPG:
	// 	fileType = entities.FileTypeJpg
	// default:
	// 	return fmt.Errorf("Unknown file type: %v", mf.Type)
	// }

	// relPath, err := im.fm.AddFile(mf.Camera.ID, f.Data, mf.Timestamp, fileType)

	// if err != nil {
	// 	im.logger.Error("Failed to save file",
	// 		zap.String("name", f.Name), zap.String("camera", mf.Camera.ID), zap.Error(err))
	// 	return fmt.Errorf("Failed to safe file %v: %w", f.FullName, err)
	// }

	// fileData, err := im.dm.AddFile(relPath, fileType, mf.Camera.ID, mf.Timestamp, mf.Duration)

	// if err != nil {
	// 	im.logger.Error("Failed to save file data", zap.Error(err),
	// 		zap.String("name", f.Name), zap.String("camera", mf.Camera.ID))
	// 	return fmt.Errorf("Failed to save file data: %w", err)
	// }

	// im.logger.Info("Ingested file", zap.String("camera", mf.Camera.ID), zap.String("path", mf.Path))

	f.Close()
	return nil
}
