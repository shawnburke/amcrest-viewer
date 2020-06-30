package cameras

import (

	"go.uber.org/fx"
	"go.uber.org/zap"

	gcommon "github.com/shawnburke/amcrest-viewer/common"
	cc "github.com/shawnburke/amcrest-viewer/cameras/common"
	"github.com/shawnburke/amcrest-viewer/cameras/amcrest"
)

var Module = fx.Options(
	fx.Provide(amcrest.New),
	fx.Provide(NewRegistry),
)


type Registry interface {
	Get(t string) (cc.Type, error)
}

type RegistryParams struct {
	fx.In
	Logger      *zap.Logger
	Bus         gcommon.EventBus
	CameraTypes   []cc.Type `group:"cameras"`
}

func NewRegistry(params RegistryParams) (Registry, error) {

	return &cameraRegistry{
		cams: params.CameraTypes,
		bus: params.Bus,
		logger: params.Logger,
	}, nil
}

type cameraRegistry struct {
	cams []cc.Type
	bus gcommon.EventBus
	logger *zap.Logger
}


func (r *cameraRegistry) Get(cameraType string) (cc.Type, error) {
	for _, c := range r.cams {
		if (c.Name() == cameraType) {
			return c, nil
		}
	}
	return nil, nil
}
