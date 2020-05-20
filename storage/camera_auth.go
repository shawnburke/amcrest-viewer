package storage

import (
	"strings"

	"go.uber.org/config"
	"go.uber.org/zap"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
)

func newCameraAuth(
	logger *zap.Logger,
	cfg config.Provider,
	data data.Repository,
) (common.Auth, error) {

	pwd := ""

	err := cfg.Get("ftp.password").Populate(&pwd)

	if err != nil {
		return nil, err
	}

	return &cameraAuth{
		storage:  data,
		logger:   logger,
		password: pwd,
	}, nil
}

type cameraAuth struct {
	storage  data.Repository
	logger   *zap.Logger
	password string
}

func (ca *cameraAuth) getCameras() ([]*entities.Camera, error) {
	return ca.storage.ListCameras()
}

func (ca *cameraAuth) IsAllowed(user, pwd string) bool {

	if ca.password != "" && ca.password != pwd {
		return false
	}

	cams, err := ca.getCameras()

	if err != nil {
		ca.logger.Error("Erorr getting cameras", zap.Error(err))
		return false
	}

	for _, cam := range cams {

		if strings.EqualFold(cam.CameraID(), user) {
			return true
		}
	}

	return false

}
