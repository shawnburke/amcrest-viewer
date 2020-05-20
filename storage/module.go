package storage

import (
	"go.uber.org/fx"

	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/file"
)

var Module = fx.Options(
	fx.Provide(newCameraAuth),
	fx.Provide(file.NewWithConfig),
	fx.Provide(data.NewFromConfig),
	fx.Provide(data.NewRepository),
)
