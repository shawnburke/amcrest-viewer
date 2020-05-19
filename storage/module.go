package storage

import (
	"strings"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Option(
	fx.Provide(newCameraAuth),
)

func newCameraAuth(logger *zap.Logger, data data.Repository) (common.Auth, error) {
	return &cameraAuth{
		storage: data,
		logger:  logger,
	}, nil
}

type cameraAuth struct {
	storage data.Repository
	logger  *zap.Logger
}

func (ca *cameraAuth) IsAllowed(user, pwd string) bool {

	cams, err := ca.storage.ListCameras()

	if err != nil {
		ca.logger.Error("Erorr getting cameras", zap.Error(err))
		return false
	}

	for _, cam := range cams {

		// TPDO: check password against config

		if strings.EqualFold(cam.CameraID(), user) {
			return true
		}
	}

	return false

}
