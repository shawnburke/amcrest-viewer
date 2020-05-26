package storage

import (
	"fmt"
	"os"
	"path"

	"github.com/jmoiron/sqlx"
	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/file"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Provide(newCameraAuth),
	fx.Provide(newFileManagerWithConfig),
	fx.Provide(newDbWithConfig),
	fx.Provide(data.NewRepository),
)

func newDbWithConfig(cfg config.Provider, p *common.Params, lifecycle fx.Lifecycle) (*sqlx.DB, error) {

	// default to using params if present
	ccfg, err := data.NewConfig(cfg)

	if err != nil {
		return nil, err
	}

	if p.DataDir != "" {
		ccfg.DSN = path.Join(p.DBDir(), "cam.db")
		err = os.MkdirAll(p.DBDir(), os.ModeDir|os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("Failed to make DB dir (%s): %w", p.DBDir(), err)
		}
	}

	return data.New(ccfg, lifecycle)
}

func newFileManagerWithConfig(logger *zap.Logger, cfg config.Provider, p *common.Params, lifecycle fx.Lifecycle) (file.Manager, error) {

	// default to using params if present
	fcfg, err := file.NewConfig(cfg)

	if err != nil {
		return nil, err
	}

	if p.DataDir != "" {
		fcfg.RootDir = p.FileDir()
		err = os.MkdirAll(p.DBDir(), os.ModeDir|os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("Failed to make file dir (%s): %w", p.DBDir(), err)
		}
	}

	return file.New(logger, fcfg)
}
