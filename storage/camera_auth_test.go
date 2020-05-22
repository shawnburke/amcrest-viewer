package storage

import (
	"testing"
	"time"

	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/stretchr/testify/require"
	"go.uber.org/config"
	"go.uber.org/zap"
)

func TestAuth(t *testing.T) {

	cfgYaml := `
ftp:
  password: foobar	
`

	provider, err := config.NewYAMLProviderFromBytes([]byte(cfgYaml))
	require.NoError(t, err)

	mr := &mockRepo{}
	camAuth, err := newCameraAuth(zap.NewNop(), provider, mr)
	require.NoError(t, err)

	allowed := camAuth.IsAllowed("magic-12", "xyz")
	require.False(t, allowed)

	allowed = camAuth.IsAllowed("magic-12", "foobar")
	require.True(t, allowed)
}

type mockRepo struct {
}

// Camera operations
func (mr *mockRepo) AddCamera(name string, t string, host *string) (*entities.Camera, error) {
	panic("not implemented") // TODO: Implement
}

func (mr *mockRepo) GetCamera(id string) (*entities.Camera, error) {
	panic("not implemented") // TODO: Implement
}

func (mr *mockRepo) DeleteCamera(id string) (bool, error) {
	panic("not implemented") // TODO: Implement
}

func (mr *mockRepo) UpdateCamera(id string, name *string, host *string, enabled *bool) (*entities.Camera, error) {
	panic("not implemented") // TODO: Implement
}

func (mr *mockRepo) SeenCamera(id string) error {
	panic("not implemented") // TODO: Implement
}

func (mr *mockRepo) ListCameras() ([]*entities.Camera, error) {
	return []*entities.Camera{
		{
			Name: "Camera 1",
			Type: "magic",
			ID:   12,
		},
	}, nil

}

// File operations
func (mr *mockRepo) AddFile(path string, t int, cameraID string, l int, timestamp time.Time, duration *time.Duration) (*entities.File, error) {
	panic("not implemented") // TODO: Implement
}

func (mr *mockRepo) GetFile(id int) (*entities.File, error) {
	panic("not implemented") // TODO: Implement
}

func (mr *mockRepo) ListFiles(cameraID string, start *time.Time, end *time.Time, fileType *int) ([]*entities.File, error) {
	panic("not implemented") // TODO: Implement
}
