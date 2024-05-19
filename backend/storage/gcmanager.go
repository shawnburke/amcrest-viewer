package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/file"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type GCManager interface {
	Start() error
	Stop() error
	Cleanup() error
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
	sync.Mutex
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

	gc.Lock()
	defer gc.Unlock()

	// get the list of cameras
	cams, err := gc.data.ListCameras()
	if err != nil {
		gc.logger.Error("Error getting camera list", zap.Error(err))
		return err
	}

	errs := []error{}

	for _, cam := range cams {
		camCount := 0
		mbCount := 0
		// get the files from the db for this camera
		cutoff := gc.time.Now().AddDate(0, 0, cam.MaxFileAgeDays*-1)

		gc.logger.Info("Running GC cleanup", zap.String("camera-id", cam.CameraID()), zap.Time("before", cutoff))

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

			found, err := gc.data.DeleteFile(file.ID)
			if err != nil {
				gc.logger.Error("Error deleting file from DB",
					zap.Int("file-id", file.ID),
					zap.String("file-path", file.Path),
					zap.Error(err))
				errs = append(errs, fmt.Errorf("error deleting file %d from disk: %w", file.ID, err))
				continue
			}

			if !found {
				gc.logger.Warn("Tried to delete non-existent file",
					zap.String("camera-id", cam.CameraID()),
					zap.Int("file-id", file.ID),
				)
			}

			_, err = gc.files.DeleteFile(file.Path)
			if err != nil {
				gc.logger.Error("Error deleting file from disk",
					zap.Int("file-id", file.ID),
					zap.String("file-path", file.Path),
					zap.Error(err),
				)
				errs = append(errs, fmt.Errorf("error deleting file %d from disk: %w", file.ID, err))
				continue
			}
			camCount++
			mbCount += (file.Length / (1024 * 1024))
		}
		gc.logger.Info("GC cleanup", zap.String("camera-id", cam.CameraID()), zap.Int("file-count", camCount), zap.Int("size-mb", mbCount))

		// TODO: size-based cleanup
	}

	if len(errs) > 0 {
		messages := ""

		for _, err := range errs {
			messages = messages + "\n" + err.Error()
		}
		return errors.New("Errors doing GC:\n" + messages)
	}

	return nil
}

func (gc *gcManager) runCleanup() {
	ticker := time.NewTicker(gc.period)

	for {

		gc.cleanupCore()
		gc.gcFiles(common.GetTempDir(), time.Hour*24)

		select {
		case <-ticker.C:
		case <-gc.done:
			gc.logger.Info("Exiting GC loop")
			return
		}
	}

}

func (gc *gcManager) gcFiles(
	dir string,
	maxAge time.Duration,
) {

	// walk the directory tree and delete any files
	// older than maxAge

	cutoff := gc.time.Now().Add(maxAge * -1)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if (dir == path) || (info == nil) {
			return nil
		}

		// Skip if not a file (i.e., is a directory)
		if info.IsDir() {
			files, err := os.ReadDir(path)
			if err == nil && len(files) == 0 {
				if err := os.RemoveAll(path); err != nil {
					gc.logger.Error("Error deleting empty temp dir", zap.String("path", path), zap.Error(err))
				}
			}
			return nil
		}

		// Delete the file if it's older than cutoff
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				gc.logger.Error("Error deleting temp file", zap.String("path", path), zap.Error(err))
			}
		}

		return nil
	})

	if err != nil {
		gc.logger.Error("Error cleaning up temp dir", zap.Error(err))
	}

}
func (gc *gcManager) Cleanup() error {
	return gc.cleanupCore()
}

func (gc *gcManager) Stop() error {
	close(gc.done)
	return nil
}
