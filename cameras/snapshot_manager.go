package cameras

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	cc "github.com/shawnburke/amcrest-viewer/cameras/common"
	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/models"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewSnapshotManager(
	logger *zap.Logger,
	data data.Repository,
	lifecycle fx.Lifecycle,
	registry Registry,
	bus common.EventBus,
	cfg config.Provider) error {

	sm := &snapshotManager{
		interval: common.SnapshotFrequency,
		data:     data,
		logger:   logger,
		registry: registry,
		bus:      bus,
		close:    make(chan struct{}),
		types:    map[string]cc.Type{},
	}

	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				return sm.start()
			},
			OnStop: func(ctx context.Context) error {
				return sm.stop()
			},
		},
	)
	return nil
}

type snapshotManager struct {
	logger   *zap.Logger
	data     data.Repository
	interval time.Duration
	registry Registry
	bus      common.EventBus
	close    chan struct{}
	types    map[string]cc.Type
}

func (sm *snapshotManager) start() error {
	go sm.run()
	return nil
}

func (sm *snapshotManager) stop() error {
	close(sm.close)
	return nil
}

func (sm *snapshotManager) shouldSnapshot(cam *entities.Camera) cc.Type {
	camType, ok := sm.types[cam.Type]

	if !ok {
		ct, err := sm.registry.Get(cam.Type)
		if err != nil {
			sm.logger.Error("Error getting camera type", zap.Error(err))
			return nil
		}
		sm.types[cam.Type] = ct
		camType = ct
	}

	if camType == nil {
		return nil
	}

	caps := camType.Capabilities()

	if caps.Snapshot {
		return camType
	}
	return nil
}

func (sm *snapshotManager) run() {
	ticker := time.NewTicker(sm.interval)
	for {

		select {
		case <-sm.close:
			return
		case <-ticker.C:

		}

		// get the cameras
		cams, err := sm.data.ListCameras()

		if err != nil {
			sm.logger.Error("Error getting cameras", zap.Error(err))
			continue
		}

		wg := &sync.WaitGroup{}

		for _, cam := range cams {

			ct := sm.shouldSnapshot(cam)
			if ct == nil {
				continue
			}

			wg.Add(1)
			go func(cam *entities.Camera) {

				defer wg.Done()

				reader, err := ct.Snapshot(cam)

				if err != nil {
					sm.logger.Error("Error getting snapshot", zap.String("camera", cam.CameraID()), zap.Error(err))
					return
				}

				if reader == nil {
					// some cameras will also FTP
					// the snapshot so we don't need to do this
					// a second time
					return
				}

				defer reader.Close()

				bytes, err := ioutil.ReadAll(reader)
				if err != nil {
					sm.logger.Error("Error reading file", zap.Error(err))
					return
				}

				ts := time.Now()
				mediaFile := &models.MediaFile{
					Name:      fmt.Sprintf("snapshot-%s-%d.jpg", cam.CameraID(), ts.Unix()),
					CameraID:  cam.CameraID(),
					Type:      models.JPG,
					Timestamp: ts,
				}

				event := storage.NewMediaFileAvailableEvent(mediaFile, bytes)

				err = sm.bus.Send(event)
				if err != nil {
					sm.logger.Error("Error sending snapshot event", zap.Error(err))
				}

			}(cam)

		}
		wg.Wait()

	}
}
