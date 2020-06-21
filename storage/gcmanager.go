package storage

import (
	"context"
	"time"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/file"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type GCManager interface {
	Start() error
	Stop() error
}

func NewGCManager(
	logger *zap.Logger,
	data data.Repository,
	files file.Manager,
	t common.Time,
	lifecycle fx.Lifecycle,
	fileConfig *file.Config,
) (GCManager, error) {

	mgr := &gcManager{
		logger:   logger,
		data:     data,
		files:    files,
		time:     t,
		period:   time.Hour * 24,
		disabled: fileConfig.GCDisabled,
	}

	if lifecycle != nil {
		lifecycle.Append(fx.Hook{
			OnStart: func(context.Context) error {
				return mgr.Start()
			},
			OnStop: func(ctx context.Context) error {
				return mgr.Stop()
			},
		})
	}

	return mgr, nil

}

type gcManager struct {
	logger   *zap.Logger
	period   time.Duration
	data     data.Repository
	files    file.Manager
	time     common.Time
	disabled bool
	done     chan (struct{})
}

func (gc *gcManager) Start() error {

	gc.done = make(chan struct{})

	go gc.runCleanup()
	return nil
}

func (gc *gcManager) cleanupCore() error {
	// get the list of cameras
	cams, err := gc.data.ListCameras()
	if err != nil {
		gc.logger.Error("Error getting camera list", zap.Error(err))
		return err
	}

	for _, cam := range cams {
		camCount := 0
		mbCount := 0
		// get the files from the db for this camera
		cutoff := gc.time.Now().AddDate(0, 0, cam.MaxFileAgeDays*-1)

		filter := &data.ListFilesFilter{
			End: &cutoff,
		}
		files, err := gc.data.ListFiles(cam.CameraID(), filter)
		if err != nil {
			gc.logger.Error("Error getting camera files", zap.String("camera-id", cam.CameraID()), zap.Error(err))
			continue
		}

		for _, file := range files {

			if gc.disabled {
				gc.logger.Info("Would delete", zap.Int("file-id", file.ID), zap.String("file-path", file.Path), zap.Time("timestamp", file.Timestamp))
				continue
			}

			err = gc.data.DeleteFile(file.ID)
			if err != nil {
				gc.logger.Error("Error deleting file from DB",
					zap.Int("file-id", file.ID),
					zap.String("file-path", file.Path),
					zap.Error(err))
			}

			_, err = gc.files.DeleteFile(file.Path)
			if err != nil {
				gc.logger.Error("Error deleting file from disk",
					zap.Int("file-id", file.ID),
					zap.String("file-path", file.Path),
					zap.Error(err),
				)
			}
			camCount++
			mbCount += (file.Length / (1024 * 1024))
		}
		gc.logger.Info("GC cleanup", zap.String("camera-id", cam.CameraID()), zap.Int("file-count", camCount), zap.Int("size-mb", mbCount))

		// TODO: size-based cleanup
	}
	return nil
}

func (gc *gcManager) runCleanup() {

	ticker := time.NewTicker(gc.period)

	for {

		select {
		case <-ticker.C:
			break
		case <-gc.done:
			gc.logger.Info("Exiting GC loop")
			return
		}

		gc.cleanupCore()
	}

}

func (gc *gcManager) Stop() error {
	close(gc.done)
	return nil
}
