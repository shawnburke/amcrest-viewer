package common

import "path"

const defaultDatadir = "data"

type Params struct {
	Host        string
	FtpPort     int
	WebPort     int
	FtpPassword string
	DataDir     string
}

func (p Params) dataDir() string {
	if p.DataDir == "" {
		return defaultDatadir
	}
	return p.DataDir
}
func (p Params) FileDir() string {

	return path.Join(p.dataDir(), "/files")
}

func (p Params) DBDir() string {
	return path.Join(p.dataDir(), "/db")
}
