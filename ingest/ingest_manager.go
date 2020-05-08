package ingest

import (
	"fmt"

	"github.com/shawnburke/amcrest-viewer/common"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Invoke(NewIngestManager),
	fx.Provide(Amcrest),
)

type IngestManagerParams struct {
	fx.In
	Logger    *zap.Logger
	Bus       common.EventBus
	Ingesters []Ingester `group:"ingester"`
}

type Ingester interface {
	Name() string
	IngestFile(f *common.File) *common.MediaFile
}

func NewIngestManager(p IngestManagerParams) error {
	im := &ingestManager{
		logger:    p.Logger,
		ingesters: map[string]Ingester{},
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
}

func (im *ingestManager) getIngesterType(user string) string {
	return amcrestIngesterType
}

func (im *ingestManager) OnEvent(e common.Event) error {
	switch ev := e.(type) {
	case *common.FileCreateEvent:
		return im.ingest(ev.File)
	case *common.FileRenameEvent:
		return im.ingest(ev.File)
	}
	return nil
}

func (im *ingestManager) ingest(f *common.File) error {
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

	// err := im.fileManager.Process(mf, f)
	// if err != nil {
	// 	im.logger.Error("Error processing file",
	// 		zap.String("file", fn.File.FullName), zap.String("type", ingesterType), zap.Error(err))
	// }

	f.Finish()
	return nil
}
