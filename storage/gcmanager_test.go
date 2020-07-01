package storage

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/file"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

func TestGC(t *testing.T) {

	dataManager := &mockDataManager{}
	fileManager := &mockFileManager{}

	m, err := NewGCManager(
		zap.NewNop(),
		dataManager,
		fileManager,
		common.NewTestTime(time.Date(2020, 06, 01, 0, 0, 0, 0, time.UTC), false),
		nil,
		&file.Config{
			DefaultTTL: time.Minute,
		},
	)

	manager := m.(*gcManager)

	require.NoError(t, err)

	err = manager.cleanupCore()
	require.NoError(t, err)

	require.Len(t, dataManager.deletedFiles, 6)
	require.Len(t, fileManager.deletedFiles, 6)

}

type mockDataManager struct {
	deletedFiles []int
}

func (dm *mockDataManager) AddCamera(name string, t string, host *string) (*entities.Camera, error) {
	panic("not implemented") // TODO: Implement
}

func (dm *mockDataManager) GetCamera(id string) (*entities.Camera, error) {
	panic("not implemented") // TODO: Implement
}

func (dm *mockDataManager) GetCameraStats(id string, start *time.Time, end *time.Time, breakdown string) (*data.CameraStats, error) {
	panic("not implemented") // TODO: Implement
}

func (dm *mockDataManager) DeleteCamera(id string) (bool, error) {
	panic("not implemented") // TODO: Implement
}

func (dm *mockDataManager) UpdateCamera(id string, name *string, host *string, enabled *bool) (*entities.Camera, error) {
	panic("not implemented") // TODO: Implement
}

func (dm *mockDataManager) UpdateCameraCreds(string, host, user, pass string) error {
	panic("not implemented") // TODO: Implement
}

func (dm *mockDataManager) SeenCamera(id string) error {
	panic("not implemented") // TODO: Implement
}

func (dm *mockDataManager) GetLatestFile(cameraID string, fileType int) (*entities.File, error) {
	panic("nyi")
}

func (dm *mockDataManager) ListCameras() ([]*entities.Camera, error) {
	return []*entities.Camera{
		{
			ID:             1,
			Name:           "Camera1",
			Type:           "amcrest",
			MaxFileAgeDays: 2,
		},
		{
			ID:             2,
			Name:           "Camera2",
			Type:           "other",
			MaxFileAgeDays: 10,
		},
	}, nil
}

// File operations
func (dm *mockDataManager) AddFile(path string, t int, cameraID string, length int, timestamp time.Time, duration *time.Duration) (*entities.File, error) {
	panic("not implemented") // TODO: Implement
}

func (dm *mockDataManager) DeleteFile(id int) error {
	dm.deletedFiles = append(dm.deletedFiles, id)
	return nil
}

func (dm *mockDataManager) GetFile(id int) (*entities.File, error) {
	panic("not implemented") // TODO: Implement
}

func (dm *mockDataManager) ListFiles(cameraID string, filter *data.ListFilesFilter) ([]*entities.File, error) {

	cam := 0

	switch cameraID {
	case "amcrest-1":
		cam = 1
	case "other-2":
		cam = 2
	default:
		panic(cameraID)
	}

	files := []*entities.File{
		{
			ID:       1,
			CameraID: cam,
			Path:     fmt.Sprintf("files/%s/1.jpg", cameraID),
		},
		{
			ID:       2,
			CameraID: cam,
			Path:     fmt.Sprintf("files/%s/2.jpg", cameraID),
		},
		{
			ID:       3,
			CameraID: cam,
			Path:     fmt.Sprintf("files/%s/3.jpg", cameraID),
		},
	}
	return files, nil
}

type mockFileManager struct {
	deletedFiles []string
}

func (fm *mockFileManager) AddFile(camera string, data []byte, timestamp time.Time, fileType int) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (fm *mockFileManager) GetFile(path string) (io.ReadCloser, error) {
	panic("not implemented") // TODO: Implement
}

func (fm *mockFileManager) GetFilePath(path string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (fm *mockFileManager) ListFiles(camera string, start *time.Time, end *time.Time, fileType *int) ([]string, error) {
	panic("not implemented") // TODO: Implement
}

func (fm *mockFileManager) DeleteFile(path string) (bool, error) {
	fm.deletedFiles = append(fm.deletedFiles, path)
	return true, nil
}

func (fm *mockFileManager) DeleteFiles(camera string, start *time.Time, end *time.Time) ([]string, error) {
	panic("not implemented") // TODO: Implement
}
