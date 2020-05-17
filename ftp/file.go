package ftp

import (
	"time"
)

// File should be renamed to FtpFile
type File struct {
	User       string
	IP         string
	Data       []byte
	Name       string
	FullName   string
	ReceivedAt time.Time
	Done       func()
}

func (f *File) Close() {
	if f != nil && f.Done != nil {
		f.Done()
	}
}
