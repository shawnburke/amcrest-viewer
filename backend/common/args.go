package common

import (
	"os"
	"path"
)

const defaultDatadir = "data"

type Params struct {
	Host        string
	FtpPort     int
	WebPort     int
	FtpPassword string
	DataDir     string
	ConfigDir   string
	FrontendDir string
}

func (p Params) dataDir() string {
	if p.DataDir == "" {
		return defaultDatadir
	}
	return p.DataDir
}

func (p Params) GetConfigDir() string {
	cd := os.Getenv("CONFIG_DIR")
	if cd != "" {
		return cd
	}

	if p.ConfigDir != "" {
		return p.ConfigDir
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd

}

func (p Params) FileDir() string {

	return path.Join(p.dataDir(), "/files")
}

func (p Params) DBDir() string {
	return path.Join(p.dataDir(), "/db")
}
